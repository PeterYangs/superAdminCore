package email

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/queue/task"
)

type EmailTask struct {
	task.BaseTask
	Parameters *Parameter
}

type Parameter struct {
	task.Parameter

	Title   string
	Email   string
	Content string
}

func NewEmailTask(title string, email string, content string) *EmailTask {

	return &EmailTask{

		BaseTask: task.BaseTask{
			TaskName: "email",
		},
		Parameters: &Parameter{
			Title:   title,
			Email:   email,
			Content: content,
		},
	}
}

func (t *EmailTask) Run() {

	fmt.Println(t.Parameters.Title, "--", t.Parameters.Email, "--", t.Parameters.Content)
}

func (t *EmailTask) BindParameters(p map[string]interface{}) {

	t.BaseTask.Bind(t.Parameters, p)

}
