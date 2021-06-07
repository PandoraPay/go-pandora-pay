package gui_non_interactive

import (
	"fmt"
	gui_interface "pandora-pay/gui/gui-interface"
)

func (g *GUINonInteractive) message(prefix string, color string, any ...interface{}) {
	text := gui_interface.ProcessArgument(any...)
	final := prefix + " " + color + " " + text

	g.writingMutex.Lock()
	fmt.Println(final)
	g.writingMutex.Unlock()
}

func (g *GUINonInteractive) Log(any ...interface{}) {
	g.message("LOG", g.colorLog, any...)
}

func (g *GUINonInteractive) Info(any ...interface{}) {
	g.message("INF", g.colorInfo, any...)
}

func (g *GUINonInteractive) Warning(any ...interface{}) {
	g.message("WARN", g.colorWarning, any...)
}

func (g *GUINonInteractive) Fatal(any ...interface{}) {
	g.message("FATAL", g.colorFatal, any...)
	panic(any)
}

func (g *GUINonInteractive) Error(any ...interface{}) {
	g.message("ERR", g.colorError, any...)
}
