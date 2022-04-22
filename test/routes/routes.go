package routes

import (
	"github.com/PeterYangs/superAdminCore/component/logs"
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/response"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/PeterYangs/superAdminCore/test/controller"
	"github.com/PeterYangs/superAdminCore/test/controller/cache"
)

func Routes(r route.Group) {

	r.Registered(route.GET, "/ping", controller.Ping).Bind()
	r.Registered(route.GET, "/test", controller.Test).Bind()
	r.Registered(route.GET, "/task", controller.Task).Bind()

	//r.Registered(route.GET, "/task", func(c *contextPlus.Context) *response.Response {
	//
	//	queue.Dispatch(email.NewEmailTask("title", "name", "content")).Delay(1 * time.Minute).Run()
	//
	//	return response.Resp().Api(1, "success", "")
	//}).Bind()

	r.Group("/cache", func(_cache route.Group) {

		_cache.Registered(route.GET, "/get", cache.Get).Bind()
		_cache.Registered(route.GET, "/put", cache.Put).Bind()
		_cache.Registered(route.GET, "/exists", cache.Exists).Bind()
		_cache.Registered(route.GET, "/remove", cache.Remove).Bind()

	})

	r.Registered(route.GET, "/ip", func(c *contextPlus.Context) *response.Response {

		return response.Resp().Api(1, "success", c.ClientIP())
	}).Bind()

	r.Registered(route.GET, "/log", func(c *contextPlus.Context) *response.Response {

		logs.NewLogs().Debug("Debug")

		return response.Resp().Api(1, "success", "")
	}).Bind()

}
