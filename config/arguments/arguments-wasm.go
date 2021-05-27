// +build wasm

package arguments

import (
	"encoding/json"
	"syscall/js"
)

const commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=name] [--set-genesis=genesis] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                             Show this screen.
  --version                             Show version.
  --testnet                             Run in TESTNET mode.
  --devnet                              Run in DEVNET mode.
  --new-devnet                          Create a new devnet genesis.
  --set-genesis=genesis                 Manually set the Genesis via a JSON. Used for devnet genesis in Browser.
  --store-wallet-type=type              Set Wallet Store Type. Accepted values: "memory|indexdb". [default: indexdb]
  --store-chain-type=type               Set Chain Store Type. Accepted values: "memory|indexdb". [default: indexdb].
  --debug                               Debug mode enabled (print log message).
  --staking                             Start staking.
  --node-name=name                      Change node name.
  --consensus=type                      Consensus type. Accepted values: "full|none". [default: full]
`

func GetArguments() []string {

	jsConfig := js.Global().Get("PandoraPayConfig")
	if jsConfig.Truthy() {
		if jsConfig.Type() != js.TypeString {
			panic("PandoraPayConfig must be an array")
		}
		out := make([]string, 0)
		str := jsConfig.String()

		err := json.Unmarshal([]byte(str), &out)
		if err != nil {
			panic(err)
		}

		return out
	}

	return nil
}
