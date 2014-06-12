package outputer

import (
	"channer"
	"service"
)

type Outputer interface {
	Output(msg interface{})
}

type OutputService interface {
	Outputer
	channer.Channer
	service.Service
}
