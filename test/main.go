package main

import (
	"github.com/PeterYangs/superAdminCore/core"
	"github.com/PeterYangs/superAdminCore/test/conf"
	"github.com/PeterYangs/superAdminCore/test/crontab"
	"github.com/PeterYangs/superAdminCore/test/routes"
)

func main() {

	c := core.NewCore()

	//加载路由
	c.LoadRoute(routes.Routes)

	//加载配置
	c.LoadConf(conf.Conf)

	//加载任务调度
	c.LoadCrontab(crontab.Crontab)

	//启动
	c.Start()

}
