package routes

import (
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/queue"
	"github.com/PeterYangs/superAdminCore/response"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/PeterYangs/superAdminCore/test/controller"
	"github.com/PeterYangs/superAdminCore/test/task/demo2"
	"time"
)

func Routes(r route.Group) {

	r.Registered(route.GET, "/ping", controller.Ping).Bind()

	r.Registered(route.GET, "/task", func(c *contextPlus.Context) *response.Response {

		queue.Dispatch(demo2.NewDemo2Task(10, "name")).Delay(1 * time.Minute).Run()

		return response.Resp().Api(1, "success", "")
	}).Bind()

}
