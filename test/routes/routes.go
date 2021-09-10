package routes

import (
	"github.com/PeterYangs/superAdminCore/route"
	"github.com/PeterYangs/superAdminCore/test/controller"
)

func Routes(r route.Group) {

	r.Registered(route.GET, "/ping", controller.Ping).Bind()

}
