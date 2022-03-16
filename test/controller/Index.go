package controller

import (
	"github.com/PeterYangs/superAdminCore/component/logs"
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/response"
)

func Ping(c *contextPlus.Context) *response.Response {

	return response.Resp().Api(1, "success", c.Session().Key())
}

func Test(c *contextPlus.Context) *response.Response {

	logs.NewLogs().Error("123").Stdout()

	return response.Resp().Api(1, "success", "")
}
