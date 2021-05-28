// +build wasm

package main

import (
	"pandora-pay/webassembly"
)

func additionalMain() {
	webassembly.Initialize(startMain)
}
