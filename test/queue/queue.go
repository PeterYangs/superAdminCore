package queue

import (
	"github.com/PeterYangs/superAdminCore/v2/queue/template"
	"github.com/PeterYangs/superAdminCore/v2/task/email"
	"github.com/PeterYangs/superAdminCore/v2/test/task/demo2"
	//namespace
)

var Queues = map[string]template.Task{
	"Demo2Task": &demo2.Demo2Task{Parameters: &demo2.Parameter{}},

	"email": &email.EmailTask{Parameters: &email.Parameter{}},

	//taskRegister
}
