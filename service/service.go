package service

import (
	"context"
	"sync"
)

type Service interface {
	Load(cxt context.Context, wait *sync.WaitGroup)
}
