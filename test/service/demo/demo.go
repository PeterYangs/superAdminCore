package demo

import (
	"context"
	"fmt"
	"time"
)

type Demo struct {
}

func NewDemo() *Demo {

	return &Demo{}
}

func (d *Demo) Load(cxt context.Context) {

	fmt.Println("加载自定义服务")

	go func() {

		defer func() {

			fmt.Println("自定义服务退出")

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
