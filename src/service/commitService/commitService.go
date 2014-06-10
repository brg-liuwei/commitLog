package commitService

import (
	"commiter"
	"errors"
	"fmt"
	"global"
	"inputer"
	"outputer"
	"reflect"
)

type CommitService struct {
	serving       bool
	inputHandler  inputer.InputService
	outputHandler outputer.OutputService
	commitHandler commiter.Commiter
}

func NewCommitService() *CommitService {
	return &CommitService{serving: false}
}

func (s *CommitService) SetInputer(inputImpl inputer.InputService) {
	s.inputHandler = inputImpl
}

func (s *CommitService) SetOutputer(outputImpl outputer.OutputService) {
	s.outputHandler = outputImpl
}

func (s *CommitService) SetCommiter(commitImpl commiter.Commiter) {
	s.commitHandler = commitImpl
}

func (s *CommitService) check() (err error) {
	switch {
	case s.inputHandler == nil:
		err = errors.New("input handler cannot be nil")
	case s.outputHandler == nil:
		err = errors.New("output handler cannot be nil")
	case s.commitHandler == nil:
		err = errors.New("commit handler cannot be nil")
	default:
		err = nil
	}
	global.Log(reflect.TypeOf(s.inputHandler))
	global.Log(reflect.TypeOf(s.outputHandler))
	global.Log(reflect.TypeOf(s.commitHandler))
	return
}

func (s *CommitService) Serve() (err error) {
	err = s.check()
	if err != nil {
		return
	}

	var inputChan, outputChan chan interface{}
	inputChan = s.inputHandler.GetChan()
	if err != nil {
		return
	}
	outputChan = s.outputHandler.GetChan()
	if err != nil {
		return
	}

	var msg interface{}
	s.serving = true
	for s.serving {
		select {
		/* Commit Msg */
		case msg = <-inputChan:
			err = s.commitHandler.Commit(msg)
			if err != nil {
				global.Log(err)
			}

		/* Get Msg by Id??? */
		case msgId := <-outputChan:
			fmt.Println("get msg id: ", msgId)
			msg, err = s.commitHandler.Consume()
			if err != nil {
				global.Log(err)
				s.outputHandler.Output(err)
			} else {
				s.outputHandler.Output(msg)
			}
		}
	}
	return
}

func (s *CommitService) Stop() {
	s.serving = false
}
