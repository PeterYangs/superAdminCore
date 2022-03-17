package main

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/cache"
	"github.com/PeterYangs/superAdminCore/core"
	"github.com/PeterYangs/superAdminCore/test/conf"
	"github.com/joho/godotenv"
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

	err := cache.Cache().Put("ppp", "ccccgg", 0)

	if err != nil {

		fmt.Println(err)

		return
	}

	fmt.Println(cache.Cache().Get("ppp"))
	fmt.Println(cache.Cache().Exists("bbb"))

}
