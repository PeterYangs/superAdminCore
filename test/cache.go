package main

import (
	"github.com/PeterYangs/superAdminCore/cache"
	"github.com/PeterYangs/superAdminCore/core"
	"github.com/PeterYangs/superAdminCore/test/conf"
	"github.com/joho/godotenv"
	"time"
)

func init() {

	err := godotenv.Load(".env")

	if err != nil {
		panic("配置文件加载失败")
	}
}

func main() {

	c := core.NewCore()

	//加载配置(这是第一步)
	c.LoadConf(conf.Conf)

	cache.Cache().Put("ppp", "pppp", 1*time.Minute)

}
