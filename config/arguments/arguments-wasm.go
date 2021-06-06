// +build wasm

package arguments

const commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=name] [--set-genesis=genesis] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type] [--tcp-max-clients=limit] [--seed-wallet-nodes-info=bool]
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
  --tcp-max-clients=limit               Change limit of clients [default: 1].
  --node-name=name                      Change node name.
  --consensus=type                      Consensus type. Accepted values: "full|wallet|none". [default: full]
  --seed-wallet-nodes-info=bool         Storing and serving additional info to wallet nodes. [default: true]. To enable, it requires full node
`
