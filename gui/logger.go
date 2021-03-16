package gui

import (
	"os"
	"path/filepath"
)

type Logger struct {
	generalLog *os.File
}

var logger = Logger{}

func InitLogger() {

	absPath, err := filepath.Abs("./_build/logs")
	if err != nil {
		panic("Error reading given path:" + err.Error())
	}

	logger.generalLog, err = os.OpenFile(absPath+"/log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Error opening file:" + err.Error())
	}

}
