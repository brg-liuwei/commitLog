package main

import (
	"fmt"
	"message"
	"net/rpc"
)

//func main() {
//	cli, err := rpc.DialHTTP("tcp", "localhost:9999")
//	if err != nil {
//		panic(err)
//	}
//	defer cli.Close()
//
//	var msg message.Msg
//	err = cli.Call("RpcHandler.GetRemoteMsg", 0, &msg)
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		fmt.Println("MessageId: ", msg.MsgId)
//		fmt.Println("MessageContent: ", msg.Content)
//	}
//}

func main() {
	cli, err := rpc.DialHTTP("tcp", "localhost:9999")
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	//for err == nil {
	for i := 0; i != 1000; i++ {
		var msg message.Msg
		err = cli.Call("RpcHandler.GetRemoteMsg", 0, &msg)
		if err != nil {
			break
		}
		fmt.Println("MessageId: ", msg.MsgId)
		fmt.Println("MessageContent: ", msg.Content)
	}
	fmt.Println("end of err: ", err)
}
