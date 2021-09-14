package register

import (
	"github.com/PeterYangs/superAdminCore/queue/task/demo"
	"github.com/PeterYangs/superAdminCore/queue/template"
	//namespace
	"sync"
)

type handles struct {
	Tasks map[string]template.Task
	Lock  sync.Mutex
}

var Register = map[string]template.Task{
	"DemoTask": &demo.DemoTask{Parameters: &demo.Parameter{}},
	//taskRegister
}

var Handles = &handles{
	Tasks: make(map[string]template.Task),
	Lock:  sync.Mutex{},
}

func (h *handles) Init() {

	h.Lock.Lock()

	defer h.Lock.Unlock()

	h.Tasks = Register

}

func (h *handles) GetTask(name string) (template.Task, bool) {

	h.Lock.Lock()

	defer h.Lock.Unlock()

	task, ok := h.Tasks[name]

	return task, ok
}
