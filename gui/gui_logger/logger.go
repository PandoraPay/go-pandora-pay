package gui_logger

import (
	"os"
	"time"
)

type GUILogger struct {
	GeneralLog *os.File
}

func CreateLogger() (*GUILogger, error) {

	logger := &GUILogger{}
	var err error

	if _, err = os.Stat("./logs"); os.IsNotExist(err) {
		if err = os.Mkdir("./logs", 0755); err != nil {
			return nil, err
		}
	}

	t := time.Now()
	filename := "log_" + t.Format("2006_01_02") + ".log"

	if logger.GeneralLog, err = os.OpenFile("./logs/"+filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666); err != nil {
		return nil, err
	}

	return logger, nil
}
