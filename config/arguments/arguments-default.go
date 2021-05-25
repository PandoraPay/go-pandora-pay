// +build !wasm

package arguments

const commands = `PANDORA PAY.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=<name>] [--tcp-server-port=<port>] [--tcp-server-address=<address>] [--tor-onion=<onion>] [--instance=<number>] [--set-genesis=<genesis>] [--store-type=<type>]
  pandorapay -h | --help

Options:
  -h --help     						Show this screen.
  --version     						Show version.
  --testnet     						Run in TESTNET mode.
  --devnet     							Run in DEVNET mode.
  --new-devnet     						Create a new devnet genesis.
  --set-genesis=<genesis>				Manually set the Genesis via a JSON. By using argument "file" it will read it via a file.
  --store-type=<type>					Set the Store Type. Accepted values: "bolt|bunt|memory"
  --debug     							Debug mode enabled (print log message).
  --staking     						Start staking
  --node-name=<name>   					Change node name
  --tcp-server-port=<port>				Change node tcp server port
  --tcp-server-address=<address>		Change node tcp address
  --tor-onion=<onion>					Define your tor onion address to be used.
  --instance=<number>					Number of forked instance (when you open multiple instances). It should me string number like "1","2","3","4" etc
`

func GetArguments() []string {
	return nil
}
