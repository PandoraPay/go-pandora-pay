// +build wasm

package arguments

import (
	"strings"
	"syscall/js"
)

func GetArguments() []string {

	jsConfig := js.Global().Get("PandoraPayConfig")
	if jsConfig.Truthy() {
		if jsConfig.Type() != js.TypeString {
			panic("PandoraPayConfig must be a string")
		}
		return strings.Split(jsConfig.String(), " ")
	}

	return nil
}
