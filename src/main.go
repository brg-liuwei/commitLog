package main

import (
	"commiter"
	"global"
	"inputer"
	"outputer"
	"service/commitService"
)

func main() {

	//global.Init("debug.log")
	global.InitStdLog()

	//inputImpl := inputer.NewZmqInputer()
	//outputImpl := outputer.NewZmqOutputer()
	//commitImpl := commiter.NewDiskCommiter()

	inputImpl := inputer.NewStdInputService()
	outputImpl := outputer.NewStdOutputService()
	commitImpl := commiter.NewMemoryCommiter()

	server := commitService.NewCommitService()
	server.SetInputer(inputImpl)
	server.SetOutputer(outputImpl)
	server.SetCommiter(commitImpl)

	go inputImpl.Serve()
	go outputImpl.Serve()

	panic(server.Serve())
}
