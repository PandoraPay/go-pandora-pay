// +build !wasm

package arguments

//use spaces for default https://github.com/docopt/docopt.go/issues/57

const commands = `PANDORA PAY.

Usage:
  pandorapay [--debugging] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=name] [--tcp-server-port=port] [--tcp-server-address=address] [--tor-onion=onion] [--instance=number] [--set-genesis=genesis] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type] [--tcp-max-client=limit] [--tcp-max-server=limit]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                             Show this screen.
  --version                             Show version.
  --testnet                             Run in TESTNET mode.
  --devnet                              Run in DEVNET mode.
  --new-devnet                          Create a new devnet genesis.
  --set-genesis=genesis                 Manually set the Genesis via a JSON. By using argument "file" it will read it via a file.
  --store-wallet-type=type              Set Wallet Store Type. Accepted values: "bolt|bunt|memory". [default: bolt]
  --store-chain-type=type               Set Chain Store Type. Accepted values: "bolt|bunt|memory".  [default: bolt]
  --debug                               Debug mode enabled (print log message).
  --staking                             Start staking.
  --node-name=name                      Change node name.
  --tcp-server-port=port                Change node tcp server port [default: 8080].
  --tcp-max-client=limit                Change limit of clients [default: 50].
  --tcp-max-server=limit                Change limit of servers [default: 500].
  --tcp-server-address=address          Change node tcp address.
  --tor-onion=onion                     Define your tor onion address to be used.
  --instance=number                     Number of forked instance (when you open multiple instances). It should me string number like "1","2","3","4" etc
  --consensus=type                      Consensus type. Accepted values: "full|wallet|none" [default: full].
`
