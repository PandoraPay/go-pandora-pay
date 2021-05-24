package arguments

import (
	"errors"
	"github.com/docopt/docopt.go"
	"pandora-pay/config"
	"pandora-pay/config/globals"
)

const commands = `PANDORA PAY.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=<name>] [--tcp-server-port=<port>] [--tcp-server-address=<address>] [--tor-onion=<onion>] [--instance=<number>] [--set-genesis=<genesis>]
  pandorapay -h | --help

Options:
  -h --help     						Show this screen.
  --version     						Show version.
  --testnet     						Run in TESTNET mode.
  --devnet     							Run in DEVNET mode.
  --new-devnet     						Create a new devnet genesis.
  --set-genesis=<genesis>				Manually set the Genesis. Used for devnet genesis. 
  --debug     							Debug mode enabled (print log message).
  --staking     						Start staking
  --node-name=<name>   					Change node name
  --tcp-server-port=<port>				Change node tcp server port
  --tcp-server-address=<address>		Change node tcp address
  --tor-onion=<onion>					Define your tor onion address to be used.
  --instance=<number>					Number of forked instance (when you open multiple instances). It should me string number like "1","2","3","4" etc
`

func InitArguments(argv []string) (err error) {

	if globals.Arguments, err = docopt.Parse(commands, argv, false, config.VERSION, false, false); err != nil {
		return errors.New("Error processing arguments" + err.Error())
	}

	return
}
