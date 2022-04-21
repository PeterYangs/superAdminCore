package queue

import (
	"fmt"
	"github.com/PeterYangs/superAdminCore/conf"
	"github.com/PeterYangs/superAdminCore/mod"
	"github.com/PeterYangs/superAdminCore/queue/register"
	"github.com/PeterYangs/tools/file/read"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"regexp"
	"runtime/debug"
)

type QueueRun struct {
}

func (q QueueRun) GetName() string {

	return "生成任务类"
}

func (q QueueRun) ArtisanRun() {

	prompt := promptui.Prompt{
		Label: "输入任务名",
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	for name, _ := range register.Register {

		if name == result {

			log.Println("该任务名已存在", string(debug.Stack()))

			return
		}

	}

	os.MkdirAll("task/"+result, 0755)

	f, err := os.OpenFile("task/"+result+"/"+result+".go", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)

	if err != nil {

		panic(err)

	}

	defer f.Close()

	script := `
package ` + result + `
	
import (	

	"github.com/PeterYangs/superAdminCore/queue/task"

)
	
type ` + capitalize(result) + `Task struct {
	task.BaseTask
	Parameters *Parameter
}
	
type Parameter struct {

	task.Parameter
	
}
	
func New` + capitalize(result) + `Task` + `() *` + capitalize(result) + `Task {
	
	return &` + capitalize(result) + `Task{
	
		BaseTask: task.BaseTask{
			TaskName: "` + result + `",
		},
			Parameters: &Parameter{
	
			},
		}
	}
	
func (t *` + capitalize(result) + `Task) Run() {
	
	
	
}
	
func (t *` + capitalize(result) + `Task) BindParameters(p map[string]interface{}) {
	
	t.BaseTask.Bind(t.Parameters, p)
	
}
	`

	_, err = f.Write([]byte(script))

	if err != nil {

		fmt.Println(err)
	}

	path := conf.Get("queue_register_path")

	if path == "" {

		panic("未找到消息队列配置路径：queue_register_path")
	}

	b, err := read.Open(path.(string)).Read()

	newScript := regexp.MustCompile("//namespace").ReplaceAllString(string(b), "\""+mod.GetModuleName()+"/task/"+result+"\"\n"+"//namespace")
	newScript = regexp.MustCompile("//taskRegister").ReplaceAllString(newScript, `"`+result+`":   &`+result+`.`+capitalize(result)+`Task{Parameters: &`+result+`.Parameter{}},`+"\n"+"//taskRegister")

	ff, err := os.OpenFile(path.(string), os.O_CREATE, 0644)

	if err != nil {

		log.Println(err, string(debug.Stack()))

		return
	}

	defer ff.Close()

	_, err = ff.Write([]byte(newScript))

	if err != nil {

		log.Println(err, string(debug.Stack()))

		return
	}

}

// Capitalize 字符首字母大写转换
func capitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				fmt.Println("Not begins with lowercase letter,")
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}
