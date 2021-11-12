package service

import "context"

type Service interface {
	Load(cxt context.Context)
}
