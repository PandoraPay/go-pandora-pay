package gui

import (
	"github.com/gizak/termui/v3/widgets"
	"os"
	"strings"
)

var logs *widgets.Paragraph

func message(color string, any ...interface{}) {
	logs.Lock()
	ss := strings.Split(logs.Text, "\n")
	pos := len(ss) - 16
	if pos < 0 {
		pos = 0
	}
	logs.Text = strings.Join(ss[pos:], "\n") + "[" + processArgument(any...) + "]" + color + "\n"
	logs.Unlock()
}

func Log(any ...interface{}) {
	message("()", any...)
}

func Info(any ...interface{}) {
	message("(fg:blue)", any...)
}

func Warning(any ...interface{}) {
	message("(fg:yellow)", any...)
}

func Fatal(any ...interface{}) error {
	message("(fg:red,fg:bold)", any...)
	os.Exit(1)
	return nil
}

func Error(any ...interface{}) error {
	message("(fg:red)", any...)
	for _, it := range any {

		switch v := it.(type) {
		case error:
			return v
		default:

		}

	}
	return nil
}

func logsInit() {
	logs = widgets.NewParagraph()
	logs.Title = "Logs"
	logs.Text = ""
	logs.WrapText = true
}
