//go:build !wasm
// +build !wasm

package start

import (
	"fmt"
	"os"
	"pandora-pay/gui"
	"pandora-pay/recovery"
)

func saveError(err error) {

	fmt.Println(err)

	file, err2 := os.Create("./error.txt")
	if err2 != nil {
		panic(err2)
	}
	defer file.Close()
	if _, err2 = file.Write([]byte(err.Error())); err2 != nil {
		panic(err2)
	}

	panic(err)
}

func startMain() {

	recovery.Safe(func() {
		if err := StartMainNow(); err != nil {
			gui.GUI.Error(err)
		}
	})

}
