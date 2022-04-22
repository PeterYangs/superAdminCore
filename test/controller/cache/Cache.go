package cache

import (
	"github.com/PeterYangs/superAdminCore/cache"
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/response"
)

func Get(c *contextPlus.Context) *response.Response {

	s, _ := cache.Cache().Get("test")

	return response.Resp().Api(1, "success", s)

}

func Put(c *contextPlus.Context) *response.Response {

	cache.Cache().Put("test", "tt", 0)

	return response.Resp().Api(1, "success", "")

}

func Exists(c *contextPlus.Context) *response.Response {

	return response.Resp().Api(1, "success", cache.Cache().Exists("test"))

}

func Remove(c *contextPlus.Context) *response.Response {

	err := cache.Cache().Remove("test")

	return response.Resp().Api(1, "success", err)

}
