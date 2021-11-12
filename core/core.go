package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/PeterYangs/gcmd2"
	"github.com/PeterYangs/superAdminCore/artisan"
	"github.com/PeterYangs/superAdminCore/component/logs"
	"github.com/PeterYangs/superAdminCore/conf"
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/crontab"
	"github.com/PeterYangs/superAdminCore/kernel"
	"github.com/PeterYangs/superAdminCore/lib/kill"
	"github.com/PeterYangs/superAdminCore/queue"
	"github.com/PeterYangs/superAdminCore/queue/register"
	"github.com/PeterYangs/superAdminCore/queue/template"
	"github.com/PeterYangs/superAdminCore/redis"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/PeterYangs/superAdminCore/service"
	"github.com/PeterYangs/tools"
	"github.com/PeterYangs/tools/file/read"
	"github.com/PeterYangs/tools/http"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
	"log"
	http_ "net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Core struct {
	Engine      *gin.Engine
	Cxt         context.Context
	Cancel      context.CancelFunc
	Wait        *sync.WaitGroup
	HttpOk      chan bool
	httpFail    chan bool
	Sigs        chan os.Signal
	Srv         *http_.Server
	Crontab     func(*crontab.Crontab)
	Routes      func(route.Group)
	Artisan     func() []artisan.Artisan
	serviceList []service.Service //启动时加载的服务
	debug       bool              //是否打开性能调试
}

var isRun = false

func NewCore() *Core {

	err := godotenv.Load(".env")

	if err != nil {
		panic("配置文件加载失败")
	}

	return &Core{}
}

// LoadRoute 加载路由
func (core *Core) LoadRoute(routes func(route.Group)) *Core {

	//route.Load(core.Engine, routes)

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

func (core *Core) Start() {

	args := os.Args
	daemon := false
	for k, v := range args {
		if v == "-d" {
			daemon = true
			args[k] = ""
		}
	}

	//直接运行则为阻塞模式，用于开发模式
	if len(args) == 1 {

		args = append(args, "block")

		core.block(args...)

		return
	}

	switch args[1] {

	case "start":

		//后台运行模式
		if daemon {

			args[1] = "block"
			core.daemonize(args...)
			return
		}

		args[1] = "block"
		core.block(args...)

	case "stop":

		err := core.stop()

		if err != nil {

			log.Println(err)

		}

	case "restart":

		err := core.stop()

		if err != nil {

			log.Println(err)

			return

		}

		args[1] = "block"
		core.daemonize(args...)

		fmt.Println("starting")

		return

	case "block":

		core.serverStart()

	case "artisan":

		core.loadService()

		artisan.RunArtisan(core.Artisan()...)

		core.Sigs <- syscall.SIGINT

		core.Wait.Wait()

	default:

		fmt.Println("命令不存在")

	}

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

func (core *Core) loadService() {

	//检测退出信号
	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//服务退出上下文，主要作用是让其他子组件协程安全退出
	cxt, cancel := context.WithCancel(context.Background())

	wait := sync.WaitGroup{}

	core.Cxt = cxt

	core.Cancel = cancel

	core.Wait = &wait

	core.Sigs = sigs

	//检测退出信号
	go core.quitCheck()

	//日志模块初始化
	core.logInit()

}

//--------------------------------------------------------

//后台运行
func (core *Core) daemonize(args ...string) {

	//后台运行模式记录重定向输出

	sysType := runtime.GOOS

	if sysType == `windows` {

		cmd := gcmd2.NewCommand(tools.Join(" ", args)+" > logs/outLog.log", context.TODO())

		err := cmd.StartNoWait()

		if err != nil {

			log.Println(err)
		}

		return
	}

	if sysType == "linux" || sysType == "darwin" {

		runUser := os.Getenv("RUN_USER")

		if runUser == "" || runUser == "nobody" {

			cmd := gcmd2.NewCommand(tools.Join(" ", args)+" > logs/outLog.log 2>&1", context.TODO())

			err := cmd.StartNoWait()

			if err != nil {

				log.Println(err)
			}

			return

		}

		//以其他用户运行服务，源命令(sudo -u nginx ./main start)
		cmd := gcmd2.NewCommand("sudo -u "+runUser+" "+tools.Join(" ", args)+" > logs/outLog.log 2>&1", context.TODO())

		err := cmd.StartNoWait()

		if err != nil {

			log.Println(err)
		}

		return
	}

	fmt.Println("平台暂不支持")

}

//阻塞运行
func (core *Core) block(args ...string) {

	sysType := runtime.GOOS

	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if sysType == `linux` {

		runUser := os.Getenv("RUN_USER")

		if runUser == "" || runUser == "nobody" {

			core.normal(args...)

			return

		}

		//以其他用户运行服务，源命令(sudo -u nginx ./main start)
		cmd := gcmd2.NewCommand("sudo -u "+runUser+" "+tools.Join(" ", args), context.TODO())

		err := cmd.Start()

		if err != nil {

			log.Println(err)

			return

		}

	}

	if sysType == `windows` {

		core.normal(args...)

		return
	}

}

//常规当前用户运行模式
func (core *Core) normal(args ...string) {

	cmd := gcmd2.NewCommand(tools.Join(" ", args), context.TODO())

	err := cmd.Start()

	if err != nil {

		log.Println(err)
	}

}

func (core *Core) stop() error {

	fmt.Println("stopping!!")

	b, err := PathExists("logs/run.pid")

	if err != nil {

		return err
	}

	if !b {

		return errors.New("run.pid文件不存在")
	}

	if b {

		pid, err := read.Open("logs/run.pid").Read()

		if err != nil {

			return err

		}

		sysType := runtime.GOOS

		var cmd *exec.Cmd

		if sysType == `windows` {

			if core.createWindowsKill() {

				cmd = exec.Command("cmd", "/c", ".\\logs\\kill.exe -SIGINT "+string(pid))
			} else {

				cmd = exec.Command("cmd", "/c", "taskkill /f /pid "+string(pid))
			}

		}

		if sysType == `linux` {

			cmd = exec.Command("bash", "-c", "kill "+string(pid))
		}

		err = cmd.Start()

		if err != nil {

			return err

		}

		err = cmd.Wait()

		if err != nil {

			return err

		}

		if sysType == `linux` {

			//等待进程退出
			for {

				time.Sleep(200 * time.Millisecond)

				wait := gcmd2.NewCommand("ps -p "+string(pid)+" | wc -l", context.TODO())

				num, waitErr := wait.CombinedOutput()

				str := strings.Replace(string(num), " ", "", -1)
				// 去除换行符
				str = strings.Replace(str, "\n", "", -1)

				if waitErr != nil {

					return waitErr

				}

				if str == "2" {

					continue

				}

				if str == "1" {

					fmt.Println("stopped!!")

					return nil
				}

			}

		}

		if sysType == `windows` {

			for {

				time.Sleep(200 * time.Millisecond)

				wait := gcmd2.NewCommand("tasklist|findstr   "+string(pid), context.TODO())

				_, waitErr := wait.CombinedOutput()

				if waitErr != nil {

					//signal.

					fmt.Println("stopped!!")

					return nil
				}

			}

		}

	}

	return nil
}

func (core *Core) serverStart() {

	//服务id生成
	kernel.IdInit()

	//检测退出信号
	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	//服务退出上下文，主要作用是让其他子组件协程安全退出
	cxt, cancel := context.WithCancel(context.Background())

	wait := sync.WaitGroup{}

	httpOk := make(chan bool)

	httpFail := make(chan bool)

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

	core.Cxt = cxt

	core.Cancel = cancel

	core.Wait = &wait

	core.HttpOk = httpOk

	core.httpFail = httpFail

	core.Sigs = sigs

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

	//加载全局中间件
	//kernel.Load()

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

	c, e := context.WithTimeout(context.Background(), 8*time.Second)

	defer e()

	if core.Srv != nil {

		//关闭http服务
		err := core.Srv.Shutdown(c)

		if err != nil {

			log.Println(err)
		}

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
				core.queueInit()

				//记录pid和启动命令
				core.runInit()

				//加载用户自定义服务
				for _, s := range core.serviceList {

					s.Load(core.Cxt)

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

func (core *Core) createWindowsKill() bool {

	b, err := PathExists("logs/kill.exe")

	if err != nil {

		return false
	}

	if b {

		return true

	} else {

		f, err := os.OpenFile("logs/kill.exe", os.O_CREATE|os.O_RDWR, 0755)

		if err != nil {

			return false
		}

		defer f.Close()

		f.Write(kill.Kill)

		return true

	}

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
