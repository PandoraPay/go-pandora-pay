// +build !wasm

package arguments

//use spaces for default https://github.com/docopt/docopt.go/issues/57

const commands = `PANDORA PAY.

Usage:
  pandorapay [--debugging] [--network=network] [--debug] [--staking] [--new-devnet] [--node-name=name] [--tcp-server-port=port] [--tcp-server-address=address] [--tor-onion=onion] [--instance=number] [--set-genesis=genesis] [--create-new-genesis=args] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type] [--tcp-max-clients=limit] [--tcp-max-server-sockets=limit] [--seed-wallet-nodes-info=bool] [--wallet-encrypt=args] [--wallet-decrypt=password] [--wallet-remove-encryption] [--wallet-derive-delegated-stake=args] [--hcaptcha-site-key=args] [--hcaptcha-secret=args] [--faucet-testnet-enabled=args] [--exit]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                                          Show this screen.
  --version                                          Show version.
  --network=network                                  Select network. Accepted values: "mainnet|testnet|devnet". [default: mainnet]
  --new-devnet                                       Create a new devnet genesis.
  --set-genesis=genesis                              Manually set the Genesis via a JSON. By using argument "file" it will read it via a file.
  --create-new-genesis=args                          Create a new Genesis. Useful for creating a new private testnet. Argument must be "0.delegatedStake,1.delegatedStake,2.delegatedStake"
  --store-wallet-type=type                           Set Wallet Store Type. Accepted values: "bolt|bunt|memory". [default: bolt]
  --store-chain-type=type                            Set Chain Store Type. Accepted values: "bolt|bunt|memory".  [default: bolt]
  --debug                                            Debug mode enabled (print log message).
  --staking                                          Start staking.
  --node-name=name                                   Change node name.
  --tcp-server-port=port                             Change node tcp server port [default: 8080].
  --tcp-max-clients=limit                            Change limit of clients [default: 50].
  --tcp-max-server-sockets=limit                     Change limit of servers [default: 500].
  --tcp-server-address=address                       Change node tcp address.
  --tor-onion=onion                                  Define your tor onion address to be used.
  --instance=number                                  Number of forked instance (when you open multiple instances). It should me string number like "1","2","3","4" etc
  --consensus=type                                   Consensus type. Accepted values: "full|wallet|none" [default: full].
  --seed-wallet-nodes-info=bool                      Storing and serving additional info to wallet nodes. [default: true]. To enable, it requires full node
  --wallet-encrypt=args                              Encrypt wallet. Argument must be "password,difficulty".
  --wallet-decrypt=password                          Decrypt wallet.
  --wallet-remove-encryption                         Remove wallet encryption.
  --wallet-derive-delegated-stake=args               Derive and export Delegated Stake. Argument must be "account,nonce,path".
  --hcaptcha-site-key=args                           hcaptcha Site key.
  --hcaptcha-secret=args                             hcaptcha Secret.
  --faucet-testnet-enabled=args                      Enable Faucet Testnet. Use "true" to enable it
  --exit                                             Exit node.
`
