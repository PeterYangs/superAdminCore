package kernel

import (
	"github.com/PeterYangs/superAdminCore/v2/contextPlus"
	"github.com/PeterYangs/superAdminCore/v2/middleware/exception"
	uuid "github.com/satori/go.uuid"
)

// Middleware 全局中间件
var Middleware []contextPlus.HandlerFunc

// Id 服务id
var Id string

func Load(list func() []contextPlus.HandlerFunc) {

	Middleware = []contextPlus.HandlerFunc{
		exception.Exception,
		//session.StartSession,
		//accessLog.AccessLog,
	}

	custom := list()

	for _, i2 := range custom {

		Middleware = append(Middleware, i2)
	}

}

func IdInit() {

	Id = uuid.NewV4().String()
}
