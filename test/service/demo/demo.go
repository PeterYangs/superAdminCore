package demo

import (
	"context"
	"fmt"
	"github.com/PeterYangs/waitTree"
	"time"
)

type Demo struct {
}

func NewDemo() *Demo {

	return &Demo{}
}

func (d *Demo) Load(cxt context.Context, wait *waitTree.WaitTree) {

	fmt.Println("加载自定义服务")

	wait.Add(1)

	go func() {

		defer func() {

			fmt.Println("自定义服务等待")

			time.Sleep(1 * time.Second)

			fmt.Println("自定义服务退出")

			wait.Done()

		}()

		for {

			select {

			case <-cxt.Done():

				return

			default:

				fmt.Println("自定义服务----")

			}

			time.Sleep(1 * time.Second)

		}

	}()

}
