package core

import (
	"context"
	"fmt"
	"github.com/PeterYangs/superAdminCore/component/logs"
	conf2 "github.com/PeterYangs/superAdminCore/conf"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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
}

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

func (core *Core) LoadConf(conf map[string]interface{}) {

	conf2.Load(conf)

}

func (core *Core) Start() {

	go core.quitCheck()

	core.logInit()

	go core.http()

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

	//if isRun {

	//删除pid文件
	_ = os.Remove("logs/run.pid")

	//}

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
