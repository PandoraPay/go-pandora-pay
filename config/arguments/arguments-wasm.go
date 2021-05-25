// +build wasm

package arguments

import (
	"strings"
	"syscall/js"
)

const commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=<name>] [--set-genesis=<genesis>] [--store-wallet-type=<type>] [--store-chain-type=<type>]
  pandorapay -h | --help

Options:
  -h --help     						Show this screen.
  --version     						Show version.
  --testnet     						Run in TESTNET mode.
  --devnet     							Run in DEVNET mode.
  --new-devnet     						Create a new devnet genesis.
  --set-genesis=<genesis>				Manually set the Genesis via a JSON. Used for devnet genesis in Browser.
  --store-wallet-type=<type>			Set Wallet Store Type. Accepted values: "memory|indexdb". By default "indexdb"".
  --store-chain-type=<type>				Set Chain Store Type. Accepted values: "memory|indexdb". By default "memory".
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
