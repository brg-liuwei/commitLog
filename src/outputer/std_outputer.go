package outputer

import (
	"fmt"
	"time"
)

/* -------------------------------------------- */
/*            StdOutputService                  */
/* -------------------------------------------- */
type StdOutputService struct {
	ch      chan interface{}
	serving bool
}

func NewStdOutputService() *StdOutputService {
	c := make(chan interface{})
	return &StdOutputService{ch: c}
}

func (s *StdOutputService) GetChan() chan interface{} {
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
