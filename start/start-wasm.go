// +build wasm

package start

import (
	"pandora-pay/webassembly"
)

func RunMain() {
	webassembly.Initialize(startMain)
}
