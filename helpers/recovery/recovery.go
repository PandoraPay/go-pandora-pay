package recovery

import (
	"pandora-pay/gui"
	"runtime/debug"
)

func SafeGo(cb func()) {
	go func() {
		Safe(cb)
	}()
}

func Safe(cb func()) {
	defer func() {
		if err := recover(); err != nil {

			stackTrace := string(debug.Stack())

			if gui.GUI != nil {
				gui.GUI.Error(err)
				gui.GUI.Error(stackTrace)
			}

		}
	}()
	cb()
}
