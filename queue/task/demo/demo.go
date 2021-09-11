package demo

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/queue/task"
)

type DemoTask struct {
	task.BaseTask
	Parameters *Parameter
}

type Parameter struct {
	task.Parameter
	Id   float64
	Name string
}

func NewDemoTask(id float64, name string) *DemoTask {

	return &DemoTask{

		BaseTask: task.BaseTask{
			TaskName: "DemoTask",
		},
		Parameters: &Parameter{
			Id:   id,
			Name: name,
		},
	}
}

func (t *DemoTask) Run() {

	fmt.Println(t.Parameters.Id, "------------", t.Parameters.Name)

}

func (t *DemoTask) BindParameters(p map[string]interface{}) {

	t.BaseTask.Bind(t.Parameters, p)

}
