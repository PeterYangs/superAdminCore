package core

import (
	conf2 "github.com/PeterYangs/superAdminCore/conf"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	http_ "net/http"
	"os"
)

type Core struct {
	Engine *gin.Engine
}

func NewCore() *Core {

	err := godotenv.Load("./.env")

	if err != nil {
		panic("配置文件加载失败")
	}

	return &Core{
		Engine: gin.Default(),
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
