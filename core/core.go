package core

import (
	"context"
	"fmt"
	"github.com/PeterYangs/gostage"
	"github.com/PeterYangs/superAdminCore/v2/artisan"
	"github.com/PeterYangs/superAdminCore/v2/conf"
	"github.com/PeterYangs/superAdminCore/v2/contextPlus"
	"github.com/PeterYangs/superAdminCore/v2/crontab"
	"github.com/PeterYangs/superAdminCore/v2/kernel"
	"github.com/PeterYangs/superAdminCore/v2/queue"
	"github.com/PeterYangs/superAdminCore/v2/queue/register"
	"github.com/PeterYangs/superAdminCore/v2/queue/template"
	"github.com/PeterYangs/superAdminCore/v2/redis"
	"github.com/PeterYangs/superAdminCore/v2/route"
	"github.com/PeterYangs/superAdminCore/v2/service"
	"github.com/PeterYangs/tools/http"
	"github.com/PeterYangs/waitTree"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
	"log"
	http_ "net/http"
	"os"
	"runtime/debug"
	"time"
)

type Core struct {
	Engine *gin.Engine
	Cxt    context.Context
	Cancel context.CancelFunc
	Wait   *waitTree.WaitTree
	HttpOk chan bool
	//httpFailCxt    context.Context
	//httpFailCancel context.CancelFunc
	//Sigs           chan os.Signal
	Srv         *http_.Server
	Crontab     func(*crontab.Crontab)
	Routes      func(route.Group)
	Artisan     func() []artisan.Artisan
	serviceList []service.Service //启动时加载的服务
	debug       bool              //是否打开性能调试
	stage       *gostage.Stage
}

func NewCore(cxt context.Context) *Core {

	err := godotenv.Load(".env")

	if err != nil {
		panic("配置文件加载失败")
	}

	stage := gostage.NewStage(cxt)

	//stage.LoadConfig(gostage.Config{
	//	RunPath: os.Getenv("RUN_USER"),
	//})

	stage.SetRunUser(os.Getenv("RUN_USER"))

	//fmt.Println(os.Getenv("RUN_USER"))

	//服务退出上下文，主要作用是让其他子组件协程安全退出
	ct, cancel := context.WithCancel(stage.GetCxt())

	return &Core{Cxt: ct, Cancel: cancel, Wait: waitTree.NewWaitTree(waitTree.Background()), stage: stage, HttpOk: make(chan bool)}
}

// LoadRoute 加载路由
func (core *Core) LoadRoute(routes func(route.Group)) *Core {

	core.Routes = routes

	return core
}

// LoadConf 加载配置
func (core *Core) LoadConf(c func() map[string]interface{}) *Core {

	conf.Load(c())

	return core
}

// LoadCrontab 加载任务调度
func (core *Core) LoadCrontab(c func(*crontab.Crontab)) *Core {

	core.Crontab = c

	return core
}

// LoadQueues 加载消息队列
func (core *Core) LoadQueues(queues map[string]template.Task) *Core {

	for s, task := range queues {

		register.Register[s] = task
	}

	return core
}

// LoadMiddleware 加载全局中间件
func (core *Core) LoadMiddleware(list func() []contextPlus.HandlerFunc) *Core {

	kernel.Load(list)

	return core
}

// LoadArtisan 加载自定义命令
func (core *Core) LoadArtisan(a func() []artisan.Artisan) *Core {

	core.Artisan = a

	return core
}

// LoadServices 加载http启动时的服务
func (core *Core) LoadServices(serv ...service.Service) *Core {

	core.serviceList = serv

	return core
}

func (core *Core) Debug() *Core {

	core.debug = true

	return core
}

func (core *Core) Start() {

	core.stage.StartFunc(func(request *gostage.Request) (string, error) {

		for {

			select {

			case <-request.GetCxt().Done():

				return "finish", nil

			default:

				core.serverStart()

			}

		}

	})

	core.stage.AddCommand("artisan", "内置命令.", func(request *gostage.Request) (string, error) {

		artisan.RunArtisan(core.Artisan()...)

		return "", nil

	}).NoConnect()

	err := core.stage.Run()

	if err != nil {

		fmt.Println(err)
	}

}

//----------------------------------------------------------------------------------------

//主进程
func (core *Core) serverStart() {

	//服务id生成
	kernel.IdInit()

	if os.Getenv("APP_DEBUG") == "true" {

		core.Engine = gin.Default()

	} else {

		core.Engine = gin.New()

	}

	if core.debug {
		//性能调试
		pprof.Register(core.Engine)
	}

	route.Load(core.Engine, core.Routes)

	//退出处理
	go core.quitDeal()

	//日志模块初始化
	//core.logInit()

	//加载子服务
	go core.boot()

	//加载http服务
	go core.http()

	//等待http服务完成
	<-core.HttpOk

	//等待其他子组件服务退出
	core.Wait.Wait()

	//fmt.Println("finish")

}

//退出处理
func (core *Core) quitDeal() {

	select {

	case <-core.Cxt.Done():

		c, can := context.WithTimeout(context.Background(), 8*time.Second)

		defer can()

		if core.Srv != nil {

			//关闭http服务
			err := core.Srv.Shutdown(c)

			if err != nil {

				log.Println(err, string(debug.Stack()))
			}

		}

	}

}

////启动日志服务
//func (core *Core) logInit() {
//
//	//日志退出标记
//	core.Wait.Add(1)
//
//	l := logs.CreateLogs()
//
//	//日志写入任务
//	go l.Task(core.Cxt, core.Wait)
//
//}

//加载子服务
func (core *Core) boot() {

	defer func() {

		core.HttpOk <- true

	}()

	//检查redis
	pingTimeoutCxt, c := context.WithTimeout(context.Background(), 1*time.Second)

	_, pingErr := redis.GetClient().Ping(pingTimeoutCxt).Result()

	c()

	if pingErr != nil {

		fmt.Println("redis连接失败，请检查")

		//发送信号让程序退出
		//core.Sigs <- syscall.SIGTERM

		core.stage.GetCancel()()

		return
	}

	client := http.Client().SetTimeout(1 * time.Second)

	for {

		select {

		//如http服务启动失败，其他子服务无需启动
		case <-core.stage.GetCxt().Done():

			fmt.Println("http启动失败")

			//发送信号让程序退出
			//core.Sigs <- syscall.SIGTERM

			return

		default:

			time.Sleep(500 * time.Millisecond)

			//验证http服务已启动完成
			str, err := client.Request().GetToString("http://127.0.0.1:" + os.Getenv("PORT") + "/_ping/" + kernel.Id)

			//fmt.Println(err, "你xx")

			//http服务启动完成后再启动子服务
			if err == nil && str == "success" {

				//开启任务调度
				if core.Crontab != nil {

					go crontab.Run(core.Wait, core.Crontab)
				}

				//队列启动
				core.queueInit()

				//记录pid和启动命令
				//core.runInit()

				//加载用户自定义服务
				for _, s := range core.serviceList {

					s.Load(core.Cxt, core.Wait)

				}

				return

			}
		}

	}

}

func (core *Core) queueInit() {

	//延迟队列的标记
	core.Wait.Add(1)

	for i := 0; i < cast.ToInt(os.Getenv("QUEUE_NUM")); i++ {

		core.Wait.Add(1)

		//启动消息队列
		go queue.Run(core.Cxt, core.Wait)

	}

}

func (core *Core) http() {

	core.Wait.Add(1)

	defer core.Wait.Done()

	srv := &http_.Server{}

	core.Srv = srv

	//设置端口
	port := os.Getenv("PORT")

	if port == "" {

		port = "8887"
	}

	srv.Addr = "127.0.0.1:" + port

	srv.Handler = core.Engine

	if err := srv.ListenAndServe(); err != nil && err != http_.ErrServerClosed {

		log.Println(err, string(debug.Stack()))

		core.stage.GetCancel()()

	}

}
