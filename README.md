# go-pandora-pay
PandoraPay blockchain in go

The main design pattern that has been taken in consideration is to be **dead-simple**. A source code that is simple is bug free and easy to be developed and improved over time.

## Running

### Running your own devnet

`go build main.go --devnet --new-devnet`

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
    - [ ] Print Wallet Homomorphic Balances
    - [X] Export JSON        
    - [ ] Import JSON        
    - [ ] Wallet Encryption
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
        - [ ] Accepting to delegate stakes from network  
- [x] Balances
    - [x] Balance and Nonce Update
    - [x] Delegating stake
    - [x] Support for Multiple Tokens   
    - [ ] Patricia Trie ?
- [ ] Homomorphic Balances
    - [ ] Homomorphic balance and nonce   
    - [ ] Multiple Tokens
    - [ ] Patricia Trie ?
- [ ] Tokens
    - [X] Token
    - [ ] Creation
    - [ ] Update  
- [ ] Transactions
    - [x] Transaction Wizard
    - [x] Transaction Builder
    - [x] Simple Transactions
    - [x] Unstake Transaction
    - [x] Delegate Stake    
    - [ ] Zether Deposit Transactions
    - [ ] Zether Withdraw Transactions
    - [ ] Zether Transfer Transactions
    - [ ] Multi Threading signature verification **
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
- [ ] API
    - [X] API blockchain explorers
    - [ ] API wallets    
- [X] Consensus
  - [X] API websockets for Forks
  - [X] Fork manager and downloader   
- [ ] Wallet
  - [ ] Webassembly build
  - [ ] Node.js and Javascript wallet
  - [ ] Signing transactions
  - [ ] Documentation **
  - [ ] Explorer

** later on optimizations and improvements

The main reasons why DPOS has been chosen over POS:
1. Delegating your stake increases security 
2. Delegating your stake to someone to stake increases anonymity as you don't need to be online for staking.
3. DPOS avoids using the griding technique to solve the POS short range vulnerability
4. Future proposals:
    1. state trie proofs to prove to light clients the state.     
    2. sharding. Creating multiple distinct shards and splitting the state trie into `n` shards
    3. creating macro blocks by selecting validators for a meta chain
  
### Debugging

Using profiling to debug memory leaks/CPU
0. Install graphviz by running `sudo apt install graphviz` 
1. use `--debugging`
2. The following command will request for a 5s CPU
   profile and will launch a browser with an SVG file. `go tool pprof -web http://:6060/debug/pprof/profile?seconds=5`
3. `go tool pprof -http :8080 http://:6060/debug/pprof/goroutine`