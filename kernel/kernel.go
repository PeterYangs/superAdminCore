package kernel

import (
	"github.com/PeterYangs/superAdminCore/contextPlus"
	"github.com/PeterYangs/superAdminCore/middleware/exception"
	uuid "github.com/satori/go.uuid"
)

// Middleware 全局中间件
var Middleware []contextPlus.HandlerFunc

// Id 服务id
var Id string

func Load() {

	Middleware = []contextPlus.HandlerFunc{
		exception.Exception,
		//session.StartSession,
		//accessLog.AccessLog,
	}

}

func IdInit() {

	Id = uuid.NewV4().String()
}
