package outputer

import (
	"errors"
	"fmt"
	"global"
	"message"
	"net"
	"net/http"
	"net/rpc"
	"reflect"
)

/* ==========================================
               RpcHandler
这里没有直接使用RpcOutputService的RPC实现，
而是另外使用一个对象来暴露RPC调用，
这样做的原因是：
RpcOutputService还有其他public成员函数，
如果直接把RpcOutputService注册到rpc中，
一是会将其他成员函数暴露出来为外部RPC使用，
二是GO对RPC函数的参数数量和格式有严格规定，
因而RpcOutputService被注册到RPC时会导致警告
=============================================*/
type RpcHandler struct {
	service *RpcOutputService
}

func NewRpcHandler(r *RpcOutputService) *RpcHandler {
	return &RpcHandler{r}
}

func (handler *RpcHandler) GetRemoteMsg(msgId int, reply *message.Msg) error {
	return handler.service.GetRemoteMsg(msgId, reply)
}

/* ================ */
/* RpcOutputService */
/* ================ */
type RpcOutputService struct {
	ch       chan interface{}
	internCh chan interface{}
	rpcAddr  string
}

func NewRpcOutputService() *RpcOutputService {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	return &RpcOutputService{ch: ch1, internCh: ch2}
}

func (r *RpcOutputService) GetChan() chan interface{} {
	return r.ch
}

func (r *RpcOutputService) SetRpcAddr(addr string) {
	r.rpcAddr = addr
}

func (r *RpcOutputService) Output(msg interface{}) {
	r.internCh <- msg
}

func (r *RpcOutputService) GetRemoteMsg(msgId int, reply *message.Msg) error {
	r.ch <- msgId
	msg := <-r.internCh

	err, ok := msg.(error)
	if ok {
		return err
	}
	_, ok = msg.(message.Msg)
	if ok {
		v := reflect.ValueOf(msg)
		/*
		   type Msg struct {
		       MsgId   uint64        <- field(0)
		       Content interface{}   <- field(1)
		   }
		*/
		reply.MsgId = v.Field(0).Uint()
		reply.Content = v.Field(1).Interface()
		return nil
	}

	fmt.Println("Reply type is: ", reflect.TypeOf(msg).Name())
	return errors.New("Unkown Type of Msg")
}

func (r *RpcOutputService) Serve() error {
	handler := NewRpcHandler(r)
	rpc.Register(handler)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", r.rpcAddr)
	if e != nil {
		panic(e)
	}
	http.Serve(l, nil)
	return nil
}

func (r *RpcOutputService) Stop() {
	global.Log("Cannot Stop Rpc Server, Use kill -9")
}
