package core

import (
	"context"
	"fmt"
	"github.com/PeterYangs/superAdminCore/component/logs"
	conf2 "github.com/PeterYangs/superAdminCore/conf"
	"github.com/PeterYangs/superAdminCore/crontab"
	"github.com/PeterYangs/superAdminCore/kernel"
	"github.com/PeterYangs/superAdminCore/redis"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/PeterYangs/tools/http"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
	"log"
	http_ "net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Core struct {
	Engine   *gin.Engine
	Cxt      context.Context
	Cancel   context.CancelFunc
	Wait     *sync.WaitGroup
	HttpOk   chan bool
	httpFail chan bool
	Sigs     chan os.Signal
	Srv      *http_.Server
	Crontab  func(*crontab.Crontab)
}

var isRun = false

func NewCore() *Core {

	err := godotenv.Load("./.env")

	if err != nil {
		panic("配置文件加载失败")
	}

	//检测退出信号
	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//服务退出上下文，主要作用是让其他子组件协程安全退出
	cxt, cancel := context.WithCancel(context.Background())

	wait := sync.WaitGroup{}

	httpOk := make(chan bool)

	httpFail := make(chan bool)

	return &Core{
		Engine:   gin.Default(),
		Cxt:      cxt,
		Cancel:   cancel,
		Wait:     &wait,
		HttpOk:   httpOk,
		httpFail: httpFail,
		Sigs:     sigs,
	}
}

// LoadRoute 加载路由
func (core *Core) LoadRoute(routes func(route.Group)) *Core {

	route.Load(core.Engine, routes)

	return core
}

// LoadConf 加载配置
func (core *Core) LoadConf(conf map[string]interface{}) *Core {

	conf2.Load(conf)

	return core
}

// LoadCrontab 加载任务调度
func (core *Core) LoadCrontab(c func(*crontab.Crontab)) *Core {

	core.Crontab = c

	return core
}

func (core *Core) Start() {

	//服务id生成
	kernel.IdInit()

	//检测退出信号
	go core.quitCheck()

	//日志模块初始化
	core.logInit()

	//加载子服务
	go core.boot()

	//加载http服务
	go core.http()

	//等待http服务完成
	<-core.HttpOk

	//等待其他子组件服务退出
	core.Wait.Wait()

	fmt.Println("finish")

}

//--------------------------------------------------------

//启动日志服务
func (core *Core) logInit() {

	//日志退出标记
	//wait.Add(1)
	core.Wait.Add(1)

	l := logs.CreateLogs()

	//日志写入任务
	go l.Task(core.Cxt, core.Wait)

}

func (core *Core) http() {

	srv := &http_.Server{}

	core.Srv = srv

	//设置端口
	port := os.Getenv("PORT")

	if port == "" {

		port = "8887"
	}

	srv.Addr = ":" + port

	srv.Handler = core.Engine

	if err := srv.ListenAndServe(); err != nil && err != http_.ErrServerClosed {

		log.Println(err)

		//http服务启动失败
		//httpFail <- true
		core.httpFail <- true

	}

}

func (core *Core) quitCheck() {

	sig := <-core.Sigs

	fmt.Println()
	fmt.Println(sig)

	if isRun {

		//删除pid文件
		_ = os.Remove("logs/run.pid")

	}

	c, e := context.WithTimeout(context.Background(), 3*time.Second)

	defer e()

	//关闭http服务
	err := core.Srv.Shutdown(c)

	if err != nil {

		log.Println(err)
	}

	//通知子组件协程退出
	//cancel()
	core.Cancel()

}

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
		core.Sigs <- syscall.SIGTERM

		return
	}

	client := http.Client().SetTimeout(1 * time.Second)

	for {

		select {

		//如http服务启动失败，其他子服务无需启动
		case <-core.httpFail:

			fmt.Println("http启动失败")

			//发送信号让程序退出
			core.Sigs <- syscall.SIGTERM

			return

		default:

			time.Sleep(200 * time.Millisecond)

			//验证http服务已启动完成
			str, err := client.Request().GetToString("http://127.0.0.1:" + os.Getenv("PORT") + "/_ping/" + kernel.Id)

			//http服务启动完成后再启动子服务
			if err == nil && str == "success" {

				//开启任务调度
				//go crontab.Run(core.Wait)
				if core.Crontab != nil {

					go crontab.Run(core.Wait, core.Crontab)
				}

				//队列启动
				//queueInit(cxt, wait)

				//记录pid和启动命令
				core.runInit()

				return

			}
		}

	}

}

func (core *Core) runInit() {

	f, err := os.OpenFile("logs/run.pid", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)

	if err != nil {

		panic(err)
	}

	//记录pid
	_, err = f.Write([]byte(cast.ToString(os.Getpid())))

	if err == nil {

		isRun = true
	}

	_ = f.Close()

}
