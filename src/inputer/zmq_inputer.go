package inputer

import (
	"global"
	zmq "zmq4"
)

type ZmqInputService struct {
	ch      chan interface{}
	serving bool
	sock    *zmq.Socket
}

func NewZmqInputService(addr string) *ZmqInputService {
	c := make(chan interface{})
	s, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		panic(err)
	}
	err = s.Connect(addr)
	if err != nil {
		panic(err)
	}
	return &ZmqInputService{ch: c, sock: s}
}

func (z *ZmqInputService) GetChan() chan interface{} {
	return z.ch
}

func (z *ZmqInputService) Input() (msg interface{}, err error) {
	msg, err = z.sock.Recv(0)
	return
}

func (z *ZmqInputService) Serve() error {
	z.serving = true
	for z.serving {
		msg, err := z.Input()
		if err != nil {
			global.Log(err)
			continue
		}
		z.ch <- msg
	}
	return nil
}

func (z *ZmqInputService) Stop() {
	z.serving = false
}
