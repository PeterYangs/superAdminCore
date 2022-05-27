package coreCopy

import (
	"context"
	"errors"
	"fmt"
	"github.com/PeterYangs/gcmd2"
	"github.com/PeterYangs/superAdminCore/v2/artisan"
	"github.com/PeterYangs/superAdminCore/v2/component/logs"
	"github.com/PeterYangs/superAdminCore/v2/conf"
	"github.com/PeterYangs/superAdminCore/v2/contextPlus"
	"github.com/PeterYangs/superAdminCore/v2/crontab"
	"github.com/PeterYangs/superAdminCore/v2/kernel"
	"github.com/PeterYangs/superAdminCore/v2/lib/kill"
	"github.com/PeterYangs/superAdminCore/v2/queue"
	"github.com/PeterYangs/superAdminCore/v2/queue/register"
	"github.com/PeterYangs/superAdminCore/v2/queue/template"
	"github.com/PeterYangs/superAdminCore/v2/redis"
	"github.com/PeterYangs/superAdminCore/v2/route"
	"github.com/PeterYangs/superAdminCore/v2/service"
	"github.com/PeterYangs/tools"
	"github.com/PeterYangs/tools/file/read"
	"github.com/PeterYangs/tools/http"
	"github.com/PeterYangs/waitTree"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
	"io"
	"log"
	http_ "net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

type Core struct {
	Engine         *gin.Engine
	Cxt            context.Context
	Cancel         context.CancelFunc
	Wait           *waitTree.WaitTree
	HttpOk         chan bool
	httpFailCxt    context.Context
	httpFailCancel context.CancelFunc
	Sigs           chan os.Signal
	Srv            *http_.Server
	Crontab        func(*crontab.Crontab)
	Routes         func(route.Group)
	Artisan        func() []artisan.Artisan
	serviceList    []service.Service //启动时加载的服务
	debug          bool              //是否打开性能调试
}

var isRun = false

func NewCore() *Core {

	err := godotenv.Load(".env")

	if err != nil {
		panic("配置文件加载失败")
	}

	//服务退出上下文，主要作用是让其他子组件协程安全退出
	cxt, cancel := context.WithCancel(context.Background())

	httpFailCxt, httpFailCancel := context.WithCancel(context.Background())

	return &Core{Cxt: cxt, Cancel: cancel, Wait: waitTree.NewWaitTree(waitTree.Background()), httpFailCxt: httpFailCxt, httpFailCancel: httpFailCancel}
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

		ok, _ := PathExists("logs/run.pid")

		if ok {

			fmt.Println("服务正在运行！")

			return
		}

		args = append(args, "block")

		core.block(args...)

		return
	}

	switch args[1] {

	case "start":

		ok, _ := PathExists("logs/run.pid")

		if ok {

			fmt.Println("服务正在运行！")

			return
		}

		//后台运行模式
		if daemon {

			args[1] = "daemon"
			core.daemonize(args...)
			return
		}

		args[1] = "block"
		core.block(args...)

	//守护进程
	case "daemon":

		sigs := make(chan os.Signal, 1)

		//退出信号
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		//记录守护进程pid
		f, err := os.OpenFile("logs/daemon.pid", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)

		if err != nil {

			panic(err)
		}

		defer func() {

			_ = os.Remove("logs/daemon.pid")

		}()

		_, err = f.Write([]byte(cast.ToString(os.Getpid())))

		_ = f.Close()

		for {

			select {

			case <-sigs:

				return

			default:

				args[1] = "block"

				core.normal(args...)

			}

		}

	case "stop":

		err := core.stop()

		if err != nil {

			log.Println(err, string(debug.Stack()))

		}

	case "restart":

		err := core.stop()

		if err != nil {

			log.Println(err, string(debug.Stack()))

			return

		}

		args[1] = "daemon"
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

	case "help":

		core.help()

	default:

		core.help()

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

	cmd := gcmd2.NewCommand(tools.Join(" ", args)+" ", context.TODO())

	if sysType == "linux" {

		cmd.SetUser(os.Getenv("RUN_USER"))

	}

	err := cmd.StartNoWaitOutErr()

	if err != nil {

		log.Println(err, string(debug.Stack()))
	}

	time.Sleep(1500 * time.Millisecond)

}

//阻塞运行
func (core *Core) block(args ...string) {

	sysType := runtime.GOOS

	//这里的退出信号是为了防止进程退出，没有其他意义
	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if sysType == `linux` {

		runUser := os.Getenv("RUN_USER")

		if runUser == "" || runUser == "nobody" {

			core.normal(args...)

			return

		}

		//以其他用户运行服务
		cmd := gcmd2.NewCommand(tools.Join(" ", args), context.TODO())

		go func(s chan os.Signal, c *gcmd2.Gcmd2) {

			select {

			case <-s:

				c.GetCmd().Process.Signal(syscall.SIGINT)

			}

		}(sigs, cmd)

		cmd.SetUser(runUser)

		core.dealOut(cmd)

		err := cmd.StartNotOut()

		if err != nil {

			log.Println(err, string(debug.Stack()))

			return

		}

	}

	if sysType == `windows` || sysType == `darwin` {

		core.normal(args...)

		return
	}

}

//常规当前用户运行模式
func (core *Core) normal(args ...string) {

	cmd := gcmd2.NewCommand(tools.Join(" ", args), context.TODO())

	core.dealOut(cmd)

	err := cmd.StartNotOut()

	if err != nil {

		log.Println(err, string(debug.Stack()))
	}

}

func (core *Core) dealOut(g *gcmd2.Gcmd2) {

	outIo, err := g.GetOutPipe()

	if err != nil {

		return
	}

	errIo, err := g.GetErrPipe()

	if err != nil {

		return
	}

	go core.out(outIo)
	go core.err(errIo)

}

func (core *Core) out(st io.ReadCloser) {

	defer st.Close()

	buf := make([]byte, 1024)

	f, fErr := os.OpenFile("logs/outLog.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if fErr != nil {

		log.Println(fErr, string(debug.Stack()))

		return
	}

	defer f.Close()

	for {

		n, readErr := st.Read(buf)

		if readErr != nil {

			if readErr == io.EOF {

				return
			}

			return
		}

		fmt.Print(string(buf[:n]))

		f.Write(buf[:n])

	}

}

func (core *Core) err(st io.ReadCloser) {

	defer st.Close()

	buf := make([]byte, 1024)

	f, fErr := os.OpenFile("logs/outErr.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)

	if fErr != nil {

		log.Println(fErr, string(debug.Stack()))

		return
	}

	defer f.Close()

	for {

		n, readErr := st.Read(buf)

		if readErr != nil {

			if readErr == io.EOF {

				return
			}

			return
		}

		fmt.Print(string(buf[:n]))

		f.Write(buf[:n])

	}
}

func (core *Core) stop() error {

	fmt.Println("stopping!!")

	_ = core.killDaemon()

	sysType := runtime.GOOS

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

				bb, waitErr := wait.CombinedOutput()

				fmt.Println(string(bb))

				if waitErr != nil {

					fmt.Println("stopped!!")

					return nil
				}

			}

		}

	}

	return nil
}

func (core *Core) killDaemon() error {

	sysType := runtime.GOOS

	d, _ := PathExists("logs/daemon.pid")

	var dCmd *exec.Cmd

	var dErr error

	if d {

		dPid, err := read.Open("logs/daemon.pid").Read()

		if err != nil {

			return err

		}

		if sysType == `windows` {

			if core.createWindowsKill() {

				dCmd = exec.Command("cmd", "/c", ".\\logs\\kill.exe -SIGINT "+string(dPid))
			} else {

				dCmd = exec.Command("cmd", "/c", "taskkill /f /pid "+string(dPid))
			}

		}

		if sysType == `linux` {

			dCmd = exec.Command("bash", "-c", "kill "+string(dPid))
		}

		dErr = dCmd.Start()

		if dErr != nil {

			return dErr

		}

		dErr = dCmd.Wait()

		if dErr != nil {

			return dErr

		}

	}

	return nil
}

//框架主进程
func (core *Core) serverStart() {

	//服务id生成
	kernel.IdInit()

	//检测退出信号
	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	httpOk := make(chan bool)

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

	core.HttpOk = httpOk

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

	srv.Addr = "127.0.0.1:" + port

	srv.Handler = core.Engine

	if err := srv.ListenAndServe(); err != nil && err != http_.ErrServerClosed {

		log.Println(err, string(debug.Stack()))

		core.httpFailCancel()

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

			log.Println(err, string(debug.Stack()))
		}

	}

	//通知子组件协程退出
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
		case <-core.httpFailCxt.Done():

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
				if core.Crontab != nil {

					go crontab.Run(core.Wait, core.Crontab)
				}

				//队列启动
				core.queueInit()

				//记录pid和启动命令
				core.runInit()

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

func (core *Core) help() {

	fmt.Println("start\t\t启动服务.")
	fmt.Println("\t\t添加-d使用守护进程.")
	fmt.Println()
	fmt.Println("stop\t\t停止服务.")
	fmt.Println()
	fmt.Println("restart\t\t重启服务.")
	fmt.Println()
	fmt.Println("artisan\t\t内置命令行.")
	fmt.Println()
	fmt.Println("help\t\t帮助.")
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
