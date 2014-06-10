package inputer

import (
	"bufio"
	"channer"
	"errors"
	"global"
	"os"
	"service"
	zmq "zmq4"
)

type Inputer interface {
	Input() (msg interface{}, err error)
}

type InputService interface {
	Inputer
	channer.Channer
	service.Service
}

type ZmqInputService struct {
	ch      chan interface{}
	serving bool
	sock    zmq.Socket
}

type StdInputService struct {
	ch      chan interface{}
	serving bool
	scanner *bufio.Scanner
}

func NewStdInputService() *StdInputService {
	c := make(chan interface{})
	s := bufio.NewScanner(os.Stdin)
	return &StdInputService{ch: c, scanner: s}
}

func (s StdInputService) GetChan() (ch chan interface{}) {
	return s.ch
}

func (s *StdInputService) Input() (msg interface{}, err error) {
	if s.scanner.Scan() {
		msg = s.scanner.Text()
	} else {
		err = errors.New("No Msg in Stdin")
	}
	return
}

func (s *StdInputService) Serve() error {
	s.serving = true
	for s.serving {
		msg, err := s.Input()
		if err != nil {
			global.Log(err)
			continue
		}
		select {
		case s.ch <- msg:
		}
	}
	return nil
}

func (s *StdInputService) Stop() {
	s.serving = false
}
