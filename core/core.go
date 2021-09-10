package core

import (
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/gin-gonic/gin"
	"log"
	http_ "net/http"
	"os"
)

type Core struct {
	Engine *gin.Engine
}

func NewCore() *Core {

	return &Core{
		Engine: gin.Default(),
	}
}

// LoadRoute 加载路由
func (core *Core) LoadRoute(routes func(route.Group)) *Core {

	route.Load(core.Engine, routes)

	return core
}

func (core *Core) Start() {

	srv := &http_.Server{}

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

	}

}
