package gui_interactive

import (
	"github.com/gizak/termui/v3/widgets"
	"pandora-pay/config"
	"pandora-pay/gui/gui_interface"
	"strings"
	"time"
)

func (g *GUIInteractive) logsRender() {
	g.logs.Lock()
	ss := strings.Split(g.logs.Text, "\n")
	pos := len(ss) - g.logs.Size().Y
	if pos < 0 {
		pos = 0
	}
	g.logs.Text = strings.Join(ss[pos:], "\n")
	g.logs.Unlock()
}

func (g *GUIInteractive) message(prefix string, color string, any ...interface{}) {

	text := gui_interface.ProcessArgument(any...)

	if config.DEBUG {
		text = time.Now().Format("2006-01-02 15:04:05  ") + text
	} else {
		text = time.Now().Format("15:04:05  ") + text
	}

	final1 := prefix + " " + text + "\n"
	final2 := "[" + text + "]" + color + "\n"

	g.logs.Lock()
	g.logger.GeneralLog.WriteString(final1)
	g.logs.Text += final2
	g.logs.Unlock()
}

func (g *GUIInteractive) Log(any ...interface{}) {
	g.message("LOG", "()", any...)
}

func (g *GUIInteractive) Info(any ...interface{}) {
	g.message("INF", "(fg:blue)", any...)
}

func (g *GUIInteractive) Warning(any ...interface{}) {
	g.message("WARN", "(fg:yellow)", any...)
}

func (g *GUIInteractive) Fatal(any ...interface{}) {
	g.message("FATAL", "(fg:red,fg:bold)", any...)
	panic(any)
}

func (g *GUIInteractive) Error(any ...interface{}) {
	g.message("ERR", "(fg:red)", any...)
}

func (g *GUIInteractive) logsInit() {
	g.logs = widgets.NewParagraph()
	g.logs.Title = "Logs"
	g.logs.Text = ""
	g.logs.WrapText = false
}
