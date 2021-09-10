package main

import (
	"github.com/PeterYangs/superAdminCore/core"
	"github.com/PeterYangs/superAdminCore/test/routes"
)

func main() {

	c := core.NewCore()

	c.LoadRoute(routes.Routes)

	c.Start()

}
