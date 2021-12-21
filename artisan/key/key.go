package key

import (
	"fmt"
	"github.com/PeterYangs/tools/file/read"
	"github.com/PeterYangs/tools/secret"
	"os"
	"regexp"
)

type Key struct {
}

func (k Key) GetName() string {

	return "生成key"
}

func (k Key) ArtisanRun() {

	_, err := os.Stat(".env")

	if err != nil {
		panic(err)
	}

	if os.IsNotExist(err) {

		fmt.Println(".env文件不存在")

		return
	}

	res, err := read.Open(".env").Read()

	if err != nil {

		panic(err)
	}

	d := secret.NewDes()

	re1 := regexp.MustCompile("KEY=[0-9A-Za-z!@#$%^&*]+").ReplaceAllString(string(res), "KEY="+string(d.GenerateKey()))

	f, err := os.OpenFile(".env", os.O_RDWR, 0644)

	if err != nil {

		panic(err)
	}

	defer f.Close()

	_, err = f.Write([]byte(re1))

	if err != nil {

		panic(err)
	}

	fmt.Println("生成成功！")

}
