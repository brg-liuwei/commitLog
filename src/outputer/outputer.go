package outputer

import (
	"channer"
	"fmt"
	"service"
	"time"
	zmq "zmq4"
)

type Outputer interface {
	Output(msg interface{})
}

type OutputService interface {
	Outputer
	channer.Channer
	service.Service
}

type ZmqOutputService struct {
	ch      chan interface{}
	serving bool
	sock    zmq.Socket
}

type StdOutputService struct {
	ch      chan interface{}
	serving bool
}

func NewStdOutputService() *StdOutputService {
	c := make(chan interface{})
	return &StdOutputService{ch: c}
}

func (s *StdOutputService) GetChan() (ch chan interface{}) {
	return s.ch
}

func (s *StdOutputService) Output(msg interface{}) {
	fmt.Println("StdOutputService output msg: ", msg)
}

func (s *StdOutputService) Serve() error {
	s.serving = true
	id := 0
	for s.serving {
		select {
		case s.ch <- id:
		default:
			time.Sleep(time.Second)
			continue
		}
		id++
	}
	return nil
}

func (s *StdOutputService) Stop() {
	s.serving = false
}
