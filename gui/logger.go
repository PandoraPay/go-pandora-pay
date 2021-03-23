package gui

import (
	"os"
	"path/filepath"
)

type Logger struct {
	generalLog *os.File
}

var logger = Logger{}

func InitLogger() (err error) {

	absPath, err := filepath.Abs("./_build/logs")
	if err != nil {
		return
	}

	logger.generalLog, err = os.OpenFile(absPath+"/log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	return nil
}
