# go-pandora-pay
PandoraPay blockchain in go

The main design pattern that has been taken in consideration is to be **dead-simple**. A source code that is simple is bug free and easy to be developed and improved over time.

### DOCS

[Installation](/docs/installation.md)

[Running](/docs/running.md)

[API](/docs/api.md)

[Scripts](/docs/scripts.md)

[Debugging](/docs/debugging.md)

[Assets](/docs/assets.md)

[Transactions](/docs/transactions.md)

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
        - [x] Integrating delegate stakes fee 
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
    - [x] Creation
    - [x] Increase Supply
    - [ ] Decrease Supply
    - [x] Liquidity Pools for Tx Fees
- [x] Transactions
    - [x] Transaction Wizard
    - [x] Transaction Builder
    - [x] Fee calculator
    - [ ] Multi Threading signature verification **   
- [x] Simple Transactions
  - [x] Unstake Tx
  - [x] Update Delegate Tx
- [x] Zether Transactions 
  - [x] Claim Tx
  - [x] Delegate Stake
  - [x] Transfer Tx
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
    - [x] P2P network
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

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.