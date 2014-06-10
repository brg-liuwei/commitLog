package global

import (
	"log"
	"os"
)

type Global struct {
	logger *log.Logger
}

var _global Global

func Init(path string) error {
	openFlag := os.O_CREATE | os.O_APPEND | os.O_RDWR
	f, err := os.OpenFile(path, openFlag, 0666)
	if err != nil {
		return err
	}
	_global.logger = log.New(f, "", log.LstdFlags)
	return nil
}

func InitStdLog() {
	_global.logger = log.New(os.Stdout, "", log.LstdFlags)
}

func Log(msg interface{}) {
	_global.logger.Println(msg)
}
