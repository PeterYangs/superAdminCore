package make

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
)

type Make struct {
}

func (m Make) GetName() string {

	return "生成命令行"
}

func (m Make) ArtisanRun() {

	prompt := promptui.Prompt{
		Label: "输入任务名",
	}

	result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	os.MkdirAll("artisan/"+result, 0755)

	f, err := os.OpenFile("artisan/"+result+"/"+result+".go", os.O_CREATE|os.O_RDWR|os.O_EXCL, 0644)

	if err != nil {

		fmt.Println(err)

		return
	}

	defer f.Close()

	str := `package ` + result + `

type ` + capitalize(result) + ` struct {
}

func (c ` + capitalize(result) + `) GetName() string {

	return "` + capitalize(result) + `"
}

func (c ` + capitalize(result) + `) ArtisanRun() {

}
`

	f.Write([]byte(str))

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
