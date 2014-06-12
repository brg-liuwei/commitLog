package inputer

import (
	"channer"
	"service"
)

type Inputer interface {
	Input() (msg interface{}, err error)
}

type InputService interface {
	Inputer
	channer.Channer
	service.Service
}
