package main

import (
	"github.com/PeterYangs/superAdminCore/core"
	"github.com/PeterYangs/superAdminCore/test/artisan"
	"github.com/PeterYangs/superAdminCore/test/conf"
	"github.com/PeterYangs/superAdminCore/test/crontab"
	"github.com/PeterYangs/superAdminCore/test/queue"
	"github.com/PeterYangs/superAdminCore/test/routes"
)

func main() {

	c := core.NewCore()

	//加载配置(这是第一步)
	c.LoadConf(conf.Conf)

	//加载路由
	c.LoadRoute(routes.Routes)

	//加载任务调度
	c.LoadCrontab(crontab.Crontab)

	//加载消息队列
	c.LoadQueues(queue.Queues)

	//加载自定义命令
	c.LoadArtisan(artisan.Artisan)

	//启动
	c.Start()

}
