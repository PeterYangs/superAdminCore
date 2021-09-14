package routes

import (
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/queue"
	"github.com/PeterYangs/superAdminCore/response"
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/PeterYangs/superAdminCore/task/email"
	"github.com/PeterYangs/superAdminCore/test/controller"
	"time"
)

func Routes(r route.Group) {

	r.Registered(route.GET, "/ping", controller.Ping).Bind()

	r.Registered(route.GET, "/task", func(c *contextPlus.Context) *response.Response {

		queue.Dispatch(email.NewEmailTask("title", "name", "content")).Delay(1 * time.Minute).Run()

		return response.Resp().Api(1, "success", "")
	}).Bind()

}
