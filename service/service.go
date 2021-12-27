package service

import (
	"context"
	"github.com/PeterYangs/waitTree"
)

type Service interface {
	Load(cxt context.Context, wait *waitTree.WaitTree)
}
