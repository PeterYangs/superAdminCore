package main

import (
	"context"
	"fmt"
	"github.com/PeterYangs/superAdminCore/v2/cache"
	"github.com/PeterYangs/superAdminCore/v2/core"
	"github.com/PeterYangs/superAdminCore/v2/test/conf"
	"github.com/joho/godotenv"
)

func init() {

	err := godotenv.Load(".env")

	if err != nil {
		panic("配置文件加载失败")
	}
}

func main() {

	c := core.NewCore(context.Background())

	//加载配置(这是第一步)
	c.LoadConf(conf.Conf)

	err := cache.Cache().Put("ppp", "789789", 0)

	if err != nil {

		fmt.Println(err)
	}

	fmt.Println(cache.Cache().Get("ppp"))

}
