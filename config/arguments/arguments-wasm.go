// +build wasm

package arguments

import (
	"strings"
	"syscall/js"
)

const commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=<name>] [--instance=<number>] [--set-genesis=<genesis>]
  pandorapay -h | --help

Options:
  -h --help     						Show this screen.
  --version     						Show version.
  --testnet     						Run in TESTNET mode.
  --devnet     							Run in DEVNET mode.
  --new-devnet     						Create a new devnet genesis.
  --set-genesis=<genesis>				Manually set the Genesis via a JSON. By using "file" argument it will read it via a file. Used for devnet genesis in Browser.
  --debug     							Debug mode enabled (print log message).
  --staking     						Start staking
  --node-name=<name>   					Change node name
`

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
