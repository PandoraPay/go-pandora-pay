// +build wasm

package arguments

const commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--debugging] [--version] [--network=network] [--debug] [--staking] [--new-devnet] [--node-name=name] [--set-genesis=genesis] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type] [--tcp-max-clients=limit] [--seed-wallet-nodes-info=bool] [--wallet-encrypt=args] [--wallet-decrypt=password] [--wallet-remove-encryption] [--wallet-derive-delegated-stake=args] [--exit]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                                          Show this screen.
  --version                                          Show version.
  --network=network                                  Select network. Accepted values: "mainnet|testnet|devnet". [default: mainnet]
  --new-devnet                                       Create a new devnet genesis.
  --set-genesis=genesis                              Manually set the Genesis via a JSON. Used for devnet genesis in Browser.
  --store-wallet-type=type                           Set Wallet Store Type. Accepted values: "memory|indexdb". [default: indexdb]
  --store-chain-type=type                            Set Chain Store Type. Accepted values: "memory|indexdb". [default: indexdb].
  --debug                                            Debug mode enabled (print log message).
  --staking                                          Start staking.
  --tcp-max-clients=limit                            Change limit of clients [default: 1].
  --node-name=name                                   Change node name.
  --consensus=type                                   Consensus type. Accepted values: "full|wallet|none". [default: full]
  --seed-wallet-nodes-info=bool                      Storing and serving additional info to wallet nodes. [default: true]. To enable, it requires full node
  --wallet-encrypt=args                              Encrypt wallet. Argument must be "password,difficulty".
  --wallet-decrypt=password                          Decrypt wallet.
  --wallet-remove-encryption                         Remove wallet encryption.
  --wallet-derive-delegated-stake=args               Derive and export Delegated Stake. Argument must be "account,nonce,path""
  --exit                                             Exit node
`
