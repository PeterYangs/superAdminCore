package main

import (
	"context"
	"fmt"
	"github.com/PeterYangs/superAdminCore/v2/cache"
	"github.com/PeterYangs/superAdminCore/v2/core"
	"github.com/PeterYangs/superAdminCore/v2/test/conf"
)

func main() {

	c := core.NewCore(context.Background())

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
