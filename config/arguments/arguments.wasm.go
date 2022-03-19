//go:build wasm
// +build wasm

package arguments

const commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--pprof] [--version] [--network=network] [--debug] [--forging] [--new-devnet] [--node-name=name] [--set-genesis=genesis] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type] [--tcp-max-clients=limit] [--seed-wallet-nodes-info=bool] [--wallet-encrypt=args] [--wallet-decrypt=password] [--wallet-remove-encryption] [--wallet-export-shared-staked-address=args] [--instance=prefix] [--instance-id=id] [--exit]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                                          Show this screen.
  --version                                          Show version.
  --network=network                                  Select network. Accepted values: "mainnet|testnet|devnet". [default: mainnet]
  --new-devnet                                       Create a new devnet genesis.
  --set-genesis=genesis                              Manually set the Genesis via a JSON. Used for devnet genesis in Browser.
  --store-wallet-type=type                           Set Wallet Store Type. Accepted values: "bunt-memory|memory|js". [default: js]
  --store-chain-type=type                            Set Chain Store Type. Accepted values: "bunt-memory|memory|js". [default: memory].
  --debug                                            Debug mode enabled (print log message).
  --forging                                          Start forging blocks.
  --instance=prefix                                  Prefix of the instance [default: 0].
  --instance-id=id                                   Number of forked instance (when you open multiple instances). It should be a string number like "1","2","3","4" etc
  --tcp-max-clients=limit                            Change limit of clients [default: 1].
  --node-name=name                                   Change node name.
  --consensus=type                                   Consensus type. Accepted values: "full|wallet|none". [default: full]
  --seed-wallet-nodes-info=bool                           Storing and serving additional info to wallet nodes. [default: true]. To enable, it requires full node
  --wallet-encrypt=args                              Encrypt wallet. Argument must be "password,difficulty".
  --wallet-decrypt=password                          Decrypt wallet.
  --wallet-remove-encryption                         Remove wallet encryption.
  --wallet-export-shared-staked-address=args                Derive and export Staked address. Argument must be "account,nonce,path".
  --exit                                             Exit node.
`
