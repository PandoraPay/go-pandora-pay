package recovery

import (
	"fmt"
	"os"
	"pandora-pay/gui"
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
			if gui.GUI != nil {
				gui.GUI.Error(err)
				gui.GUI.Close()
			}

			fmt.Println("Error: \n\n", err)
			os.Exit(0)
		}
	}()
	cb()
}
