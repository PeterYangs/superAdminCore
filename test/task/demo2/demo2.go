package demo2

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/v2/queue/task"
)

type Demo2Task struct {
	task.BaseTask
	Parameters *Parameter
}

type Parameter struct {
	task.Parameter
	Id   float64
	Name string
}

func NewDemo2Task(id float64, name string) *Demo2Task {

	return &Demo2Task{

		BaseTask: task.BaseTask{
			TaskName: "Demo2Task",
		},
		Parameters: &Parameter{
			Id:   id,
			Name: name,
		},
	}
}

func (t *Demo2Task) Run() error {

	fmt.Println(t.Parameters.Id, "------------", t.Parameters.Name)

	return nil

}

func (t *Demo2Task) BindParameters(p map[string]interface{}) {

	t.BaseTask.Bind(t.Parameters, p)

}
