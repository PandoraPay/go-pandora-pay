// +build wasm

package arguments

import (
	"strings"
	"syscall/js"
)

func init_arguments(argv []string) []string {

	jsConfig := js.Global().Get("PandoraPayConfig")
	if jsConfig.Truthy() {
		if jsConfig.Type() != js.TypeString {
			panic("PandoraPayConfig must be a string")
		}
		argv = strings.Split(jsConfig.String(), " ")
	}

	return argv
}
