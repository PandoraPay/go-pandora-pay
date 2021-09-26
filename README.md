# go-pandora-pay
PandoraPay blockchain in go

The main design pattern that has been taken in consideration is to be **dead-simple**. A source code that is simple is bug free and easy to be developed and improved over time.

## Installing

Go 1.18 will be required once released as go-pandora-pay will use generics.

Tested with 1.17 and 1.16

1. Install golang https://golang.org/doc/install
2. Installing missing packages `go get -t .`
3. Run the node

## Running

### Running your own devnet

`--debugging --network="devnet" --new-devnet --tcp-server-port="5231" --set-genesis="file"  --staking`

#### Activating devnet faucet

`--hcaptcha-site-key="10000000-ffff-ffff-ffff-000000000001" --hcaptcha-secret="0x0000000000000000000000000000000000000000" --faucet-testnet-enabled="true"`
you can also create an account on hcaptcha 

### Running the node as Tor Hidden Server
1. Install Tor
2. Configure Tor 
    - Ubuntu: 
        - open `sudo nano /etc/tor/torrc` 
        - add & save
            ``` 
            HiddenServiceDir /var/lib/tor/pandora_pay_hidden_service/
            HiddenServicePort 80 127.0.0.1:8080
            ```      
        - restart tor `sudo service tor restart`
        - copy your onion address `sudo nano /var/lib/tor/pandora_pay_hidden_service/` 
        - use the parameter `--tor-onion="YOUR_ONION_ADDRESS_FROM_ABOVE"`

## Status of Blockchain implementation:

- [x] Simple GUI
- [x] CLI commands
- [x] ECDSA
    - [x] Private Key
    - [x] Public Address (amount and paymentId)
    - [x] HD Wallet
- [x] Commit/Rollback Database
- [x] Wallet
    - [x] Save and Load
    - [x] Print Wallet Simple Balances
    - [x] Print Wallet Homomorphic Balances
    - [X] Export Address JSON        
    - [X] Import Address JSON        
    - [X] Wallet Encryption
- [x] Merkle Tree
- [x] Block
    - [x] Serialization
    - [x] Deserialization
    - [x] Hashing
    - [x] Kernel Hash
    - [x] Forger signing  
- [x] Blockchain
    - [x] Saving state
    - [x] Locking mechanism
    - [x] Difficulty Adjustment
    - [x] Timestamp maximum drift    
- [x] Forging
    - [x] Forging with wallets Multithreading    
    - [X] Forging with delegated stakes
        - [x] Accepting to delegate stakes from network  
- [x] Balances
    - [x] Balance and Nonce Update
    - [x] Delegating stake
    - [x] Support for Multiple Assets
- [x] Homomorphic Balances
    - [x] Homomorphic balance and nonce   
    - [x] Multiple Assets
- [ ] Patricia Trie ? **
- [ ] Assets
    - [X] Asset
    - [ ] Creation
    - [ ] Update  
    - [ ] Liquidity Pools for Tx Fees
- [x] Transactions
    - [x] Transaction Wizard
    - [x] Transaction Builder
    - [x] Fee calculator
    - [ ] Multi Threading signature verification **   
- [x] Simple Transactions 
  - [x] Simple Transactions
  - [x] Unstake Transaction
  - [x] Update Delegate
- [x] Zether Transactions 
  - [ ] Reclaim Transaction
  - [ ] Delegate Stake
  - [x] Transfer Transactions
- [ ] Mem Pool
    - [ ] Saving/Loading
    - [X] Inserting Txs
    - [x] Sorting by fee per byte
    - [x] Network propagation
- [X] Network
    - [X] HTTP server    
    - [X] HTTP websocket server
    - [x] HTTP websocket client
    - [X] TOR Integration
    - [ ] P2P network
- [ ] API
    - [X] API blockchain explorers
    - [ ] API wallets    
- [X] Consensus
  - [X] API websockets for Forks
  - [X] Fork manager and downloader
- [X] Webassembly build
  - [X] GUI
  - [X] Store
  - [X] Websockets
  - [X] Network
- [x] Wallet
  - [X] Node.js and Javascript wallet
  - [X] Signing transactions
  - [ ] Documentation **
  - [X] Explorer

** later on optimizations and improvements

The main reasons why DPOS has been chosen over POS:
1. Delegating your stake increases security. 
2. Delegating your stake to someone to stake increases anonymity as you don't need to be online for staking. 
3. Completely offline can be done to increase the security. 
4. DPOS avoids using the griding technique to solve the POS short range vulnerability
5. Future proposals:
    1. state trie proofs to prove to light clients the state.     
    2. sharding. Creating multiple distinct shards and splitting the state trie into `n` shards
    3. creating macro blocks by selecting specific nodes for a meta chain. This allows light consensus.
  
### Debugging

Using profiling to debug memory leaks/CPU
0. Install graphviz by running `sudo apt install graphviz` 
1. use `--debugging`
2. The following command will request for a 5s CPU
   profile and will launch a browser with an SVG file. `go tool pprof -web http://:6060/debug/pprof/profile?seconds=5`
3. `go tool pprof -http :8080 http://:6060/debug/pprof/goroutine`

#### Debugging races
 GORACE="log_path=/PandoraPay/pandora-pay-go/report" go run -race main.go 

### DOCS
[WebAssembly DOCS](/webassembly/webassembly.md)

### Scripts
`scripts/compile-wasm.sh` compiles to WASM

### Checking and Installing a specific go version
1. `go env GOROOT`
2. download from https://golang.org/doc/manage-install

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.