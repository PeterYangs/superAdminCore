package main

import (
	"context"
	"github.com/PeterYangs/superAdminCore/v2/contextPlus"
	"github.com/PeterYangs/superAdminCore/v2/core"
	"github.com/PeterYangs/superAdminCore/v2/middleware/session"
	"github.com/PeterYangs/superAdminCore/v2/test/artisan"
	"github.com/PeterYangs/superAdminCore/v2/test/conf"
	"github.com/PeterYangs/superAdminCore/v2/test/crontab"
	"github.com/PeterYangs/superAdminCore/v2/test/queue"
	"github.com/PeterYangs/superAdminCore/v2/test/routes"
)

func main() {

	c := core.NewCore(context.Background())

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

	c.LoadMiddleware(func() []contextPlus.HandlerFunc {

		return []contextPlus.HandlerFunc{
			session.StartSession,
		}
	})

	//加载自定义服务
	//c.LoadServices(demo.NewDemo())

	//一个接收全局退出信号的例子
	//go func() {
	//
	//	c.Wait.Add(1)
	//
	//	defer c.Wait.Done()
	//
	//	select {
	//	case <-c.Cxt.Done():
	//
	//		time.Sleep(3 * time.Second)
	//
	//		fmt.Println("测试退出")
	//
	//		return
	//
	//	}
	//
	//}()

	//启动
	c.Start()

}
