package gui

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"os"
	"strings"
)

var logs *widgets.Paragraph

func logsRender() {
	logs.Lock()
	ss := strings.Split(logs.Text, "\n")
	pos := len(ss) - logs.Size().Y
	if pos < 0 {
		pos = 0
	}
	logs.Text = strings.Join(ss[pos:], "\n")
	logs.Unlock()
	ui.Render(logs)
}

func message(prefix string, color string, any ...interface{}) {
	text := processArgument(any...)

	logs.Lock()
	logger.generalLog.WriteString(prefix + " " + text + "\n")
	logs.Text = logs.Text + "[" + text + "]" + color + "\n"
	logs.Unlock()

}

func Log(any ...interface{}) {
	message("LOG", "()", any...)
}

func Info(any ...interface{}) {
	message("INF", "(fg:blue)", any...)
}

func Warning(any ...interface{}) {
	message("WARN", "(fg:yellow)", any...)
}

func Fatal(any ...interface{}) {
	message("FATAL", "(fg:red,fg:bold)", any...)
	os.Exit(1)
}

func Error(any ...interface{}) {
	message("ERR", "(fg:red)", any...)
}

func logsInit() {
	logs = widgets.NewParagraph()
	logs.Title = "Logs"
	logs.Text = ""
	logs.WrapText = false
}
