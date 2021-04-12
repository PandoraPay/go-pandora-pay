package gui

import (
	"os"
	"time"
)

type Logger struct {
	generalLog *os.File
}

var logger = Logger{}

func InitLogger() (err error) {

	if _, err = os.Stat("./logs"); os.IsNotExist(err) {
		if err = os.Mkdir("./logs", 0755); err != nil {
			return
		}
	}

	t := time.Now()
	filename := "log_" + t.Format("2006_01_02") + ".log"

	logger.generalLog, err = os.OpenFile("./logs/"+filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	return nil
}
