package demo

import "fmt"

type Demo struct {
}

func (d Demo) GetName() string {

	return "自定义命令"
}

func (d Demo) ArtisanRun() {

	fmt.Println("自定义命令")

}
