//go:build !wasm
// +build !wasm

package arguments

//use spaces for default https://github.com/docopt/docopt.go/issues/57

const commands = `PANDORA PAY.

Usage:
  pandorapay [--pprof] [--network=network] [--debug] [--forging] [--new-devnet] [--run-testnet-script] [--node-name=name] [--tcp-server-port=port] [--tcp-server-address=address] [--tcp-server-auto-tls-certificate] [--tcp-server-tls-cert-file=path] [--tcp-server-tls-key-file=path] [--tor-onion=onion] [--instance=prefix] [--instance-id=id] [--set-genesis=genesis] [--create-new-genesis=args] [--store-wallet-type=type] [--store-chain-type=type] [--consensus=type] [--tcp-max-clients=limit] [--tcp-max-server-sockets=limit] [--seed-wallet-nodes-info=bool] [--wallet-encrypt=args] [--wallet-decrypt=password] [--wallet-remove-encryption] [--wallet-export-shared-staked-address=args] [--wallet-import-secret-mnemonic=mnemonic] [--wallet-import-secret-entropy=entropy] [--hcaptcha-secret=args] [--faucet-testnet-enabled=args] [--delegator-enabled=bool] [--delegator-require-auth=bool] [--delegates-maximum=args] [--auth-users=args] [--light-computations] [--delegator-fee=fee] [--delegator-reward-collector-pub-key=pubKey] [--delegator-accept-custom-keys=bool] [--exit] [--skip-init-sync]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                                          Show this screen.
  --version                                          Show version.
  --network=network                                  Select network. Accepted values: "mainnet|testnet|devnet". [default: mainnet]
  --new-devnet                                       Create a new devnet genesis.
  --run-testnet-script                               Run testnet script which will create dummy transactions in the network.
  --set-genesis=genesis                              Manually set the Genesis via a JSON. By using argument "file" it will read it via a file.
  --create-new-genesis=args                          Create a new Genesis. Useful for creating a new private testnet. Argument must be "0.stake,1.stake,2.stake"
  --store-wallet-type=type                           Set Wallet Store Type. Accepted values: "bolt|bunt|bunt-memory|memory". [default: bolt]
  --store-chain-type=type                            Set Chain Store Type. Accepted values: "bolt|bunt|bunt-memory|memory".  [default: bolt]
  --debug                                            Debug mode enabled (print log message).
  --forging                                          Start Forging blocks.
  --node-name=name                                   Change node name.
  --instance=prefix                                  Prefix of the instance [default: 0].
  --instance-id=id                                   Number of forked instance (when you open multiple instances). It should be a string number like "1","2","3","4" etc
  --tcp-server-port=port                             Change node tcp server port [default: 8080].
  --tcp-max-clients=limit                            Change limit of clients [default: 50].
  --tcp-max-server-sockets=limit                     Change limit of servers [default: 500].
  --tcp-server-address=address                       Change node tcp address.
  --tcp-server-auto-tls-certificate                  If no certificate.crt is provided, this option will generate a valid TLS certificate via autocert package. You still need a valid domain provided and set --tcp-server-address.
  --tcp-server-tls-cert-file=path                    Load TLS certificate file from given path.
  --tcp-server-tls-key-file=path                     Load TLS ke file from given path.
  --tor-onion=onion                                  Define your tor onion address to be used.
  --consensus=type                                   Consensus type. Accepted values: "full|wallet|none" [default: full].
  --seed-wallet-nodes-info=bool                      Storing and serving additional info to wallet nodes. [default: true]. To enable, it requires full node
  --wallet-import-secret-mnemonic=mnemonic           Import Wallet from a given Mnemonic. It will delete your existing wallet. 
  --wallet-import-secret-entropy=entropy             Import Wallet from a given Entropy. It will delete your existing wallet.
  --wallet-encrypt=args                              Encrypt wallet. Argument must be "password,difficulty".
  --wallet-decrypt=password                          Decrypt wallet.
  --wallet-remove-encryption                         Remove wallet encryption.
  --wallet-export-shared-staked-address=args         Derive and export Staked address. Argument must be "account,nonce,path".
  --hcaptcha-secret=args                             hcaptcha Secret.
  --faucet-testnet-enabled=args                      Enable Faucet Testnet. Use "true" to enable it
  --delegator-enabled=bool                           Enable Delegator. Will allow other users to Delegate to the node. Use "true" to enable it
  --delegator-require-auth=bool                      Delegator will require authentication.
  --delegates-maximum=args                           Maximum number of Delegates
  --delegator-fee=fee                                Fee required for Delegates
  --delegator-reward-collector-pub-key=pubKey          Delegator Reward Collector Address
  --delegator-accept-custom-keys=bool                Delegator accept custom private keys for delegated stakes. This should not be allowed in pools where the reward is split.
  --auth-users=args                                  Credential for Authenticated Users. Arguments must be a JSON "[{'user': 'username', 'pass': 'secret'}]".
  --light-computations                               Reduces the computations for a testnet node.
  --exit                                             Exit node.
  --skip-init-sync                                   Skip sync wait at when the node started. Useful when creating a new testnet.
`
