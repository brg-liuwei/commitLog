package main

import (
	"commiter"
	"global"
	"inputer"
	"os"
	"outputer"
	"service/commitService"
)

func main() {

	if len(os.Args) != 2 {
		println("Usage: ", os.Args[0], " <path>")
		return
	}
	//global.Init("debug.log")
	global.InitStdLog()

	//inputImpl := inputer.NewStdInputService()
	//outputImpl := outputer.NewStdOutputService()
	//commitImpl := commiter.NewMemoryCommiter()

	inputImpl := inputer.NewZmqInputService("tcp://localhost:9998")
	outputImpl := outputer.NewRpcOutputService()
	outputImpl.SetRpcAddr(":9999")
	//commitImpl := commiter.NewMemoryCommiter()
	commitImpl := commiter.NewDiskCommiter(os.Args[1])

	server := commitService.NewCommitService()
	server.SetInputer(inputImpl)
	server.SetOutputer(outputImpl)
	server.SetCommiter(commitImpl)

	go inputImpl.Serve()
	go outputImpl.Serve()

	panic(server.Serve())
}
