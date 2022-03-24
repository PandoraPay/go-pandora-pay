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
    - [x] Menomic Seed
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
    - [X] Forging with staked accounts
        - [x] Accepting to stakes from network
- [x] Balances
    - [x] Balance and Nonce Update
    - [x] Liquidity fee
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
    - [x] Multi Threading signature verification
- [x] Simple Transactions
  - [x] Fee calculator
  - [x] Update Asset Fee Liquidity
- [x] Zether Transactions
  - [x] Transfer
  - [x] Spend Tx
  - [x] Staking
  - [x] Staking Reward
  - [x] Asset Create
  - [x] Asset Supply Increase
  - [x] Plain Account Fund
- [ ] Mem Pool
    - [ ] Saving/Loading **
    - [X] Inserting Txs
    - [x] Sorting by fee per byte
    - [x] Network propagation
- [X] Network
    - [X] HTTP server    
    - [X] HTTP websocket server
    - [x] HTTP websocket client
    - [X] TOR Integration
    - [x] P2P network
- [x] API
    - [X] API blockchain explorers
    - [x] API wallets    
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

### Consensus UPPOS

The initial version of the consensus was DPOS, but it required all delegated accounts to have public balances and be plain accounts. Later, we switched the consensus from DPOS to a novel consensus we designed named UPPOS (Unspendable Private Proof of Stake), which is a proof of stake consensus with confidential amounts and ring signatures. The block forger will prove that he has a balance >= value T (with T being public). For cold staking or even more security, the private key of the account can be "shared" with a delegator full node. Thus, the node will generate special zether transactions that proves that the shared account has a balance >= T. To avoid the delegator steal the coins, accounts will need to have a special SpendPublicKey attached. Only the real owner of the account will know the private key of this SpendKey and can move coins out his account.    

The main reasons why UPPOS has been chosen over POS:

  1. Sharing your stake to a third party node increases security as your wallet can be secured with a cold spend private key. 
  2. Sharing your stake to someone to stake increases the privacy as you don't need to be online for staking. 
  3. Completely offline can be done to increase the security. 
  4. Griding technique and short range attack vector attack is solved using rollover window for staked accounts requiring them to get into a pending queue.

### Future proposals
  1. state trie proofs to prove to light clients the state.
  2. creating macro blocks by selecting specific nodes for a meta chain. This allows light consensus.
  3. scalability. There will be research done to understand the best way to scale up the technology.

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.