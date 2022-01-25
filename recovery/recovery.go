package recovery

import (
	"fmt"
	"os"
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
		err := recover()
		if err != nil {

			stackTrace := string(debug.Stack())

			if gui.GUI != nil {
				gui.GUI.Error(err)
				gui.GUI.Error(stackTrace)
			}

			fmt.Println("Error: \n\n", err, stackTrace)
			os.Exit(1)
		}
	}()
	cb()
}
