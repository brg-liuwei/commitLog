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

	if len(os.Args) < 2 {
		println("Usage: ", os.Args[0], " <path> [r]")
		return
	}
	global.Init("debug.log")
	//global.InitStdLog()

	inputImpl := inputer.NewZmqInputService("tcp://localhost:9998")
	outputImpl := outputer.NewRpcOutputService()
	outputImpl.SetRpcAddr(":9999")

	var commitImpl commiter.Commiter
	if len(os.Args) == 3 && os.Args[2] == "r" {
		commitImpl = commiter.RecoverDiskCommiter(os.Args[1])
	} else {
		commitImpl = commiter.NewDiskCommiter(os.Args[1])
	}

	server := commitService.NewCommitService()
	server.SetInputer(inputImpl)
	server.SetOutputer(outputImpl)
	server.SetCommiter(commitImpl)

	go inputImpl.Serve()
	go outputImpl.Serve()

	panic(server.Serve())
}
