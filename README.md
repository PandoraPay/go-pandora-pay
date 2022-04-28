# WebDollar2

WebDollar2 is a fork of PandoraPay in go

The main design pattern that has been taken in consideration is to be **dead-simple**. A source code that is simple is
bug free and easy to be developed and improved over time.

# Progress

Under development. Not working right now.

- [X] Removing Zether
- [X] Integration of ED25519
- [X] Integration of BIP32 for ED25519
- [X] Using PublicKeyHashes instead of PublicKeys
- [X] Removing Address (stakable, spendPublicKey, registration)
- [X] Creating Testnet Genesis
- [ ] DPOS contracts
- [X] DPOS consensus
- [X] Staking
- [X] WebDollar Address Format
- [X] Multi Transfer
- [ ] Web Wallet

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
    - [x] Public Address (amount and paymentID)
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
- [ ] Patricia Trie ? **
- [ ] Assets
    - [X] Asset
    - [x] Creation
    - [x] Increase Supply
    - [ ] Decrease Supply
- [x] Transactions
    - [x] Transaction Wizard
    - [x] Transaction Builder
    - [x] Fee calculator
    - [x] Multi Threading signature verification
- [x] Simple Transactions
    - [x] Fee calculator
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

The initial version of the consensus was DPOS, but it required all delegated accounts to have the balances public and
known. Later, we switched the consensus from DPOS to a novel consensus we designed named UPPOS (Unspendable Private
Proof of Stake), which is a proof of stake consensus with confidential amounts and ring signatures.

The block forger will prove that he has owns coins >= value B (with B being public) that solves the staking
equation. `Hash( PrevBlockKernelHash + StakingNonce + Timestamp) / B <= Target`

where `B` is the balance and could be between (0..Account Balance]. By `B` we understand the minimum value that
satisfies the above equation. StakingNonce is unique per address for each new block.

For cold staking or even more security, the private key of the account can be "shared" with a delegator full node. Thus,
the node will generate the special zether transactions that prove that the shared account has a balance >= B. To avoid
the delegator steal the coins, accounts will need to have a special SpendPublicKey attached. Only the real owner of the
account will know the private key of this SpendKey allowing only him to move the coins out his account. The only disadvantage is
that when a SpendKey is attached, and a user transfer his coins, it is quite fairly guessable that he is the sender from
the sender ring. A third party online service provider that behaves like a Two Factor Authenticator (even multisig) could be later
hosted by the community members and the SpendPublicKey could be the same and used by multiple people. Having multiple
addresses having the same SpendPublicKey will allow this way the Private Unspendable Accounts.

The main reasons why UPPOS has been chosen over POS:

1. Sharing your stake to a third party node increases security as your wallet can be secured with a cold spend private
   key.
2. Sharing your stake to someone to stake increases the privacy as you don't need to be online for staking.
3. Completely offline can be done to increase the security.
4. Griding technique and short range attack vector attack is solved using rollover window for staked accounts requiring
   them to get into a pending queue.

### Future proposals

1. state trie proofs to prove to light clients the state.
2. creating macro blocks by selecting specific nodes for a meta chain. This allows light consensus.
3. scalability. There will be research done to understand the best way to scale up the technology.

# DISCLAIMER:

This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p
network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO
WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to
criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.