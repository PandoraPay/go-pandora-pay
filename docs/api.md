# PandoraPay API

## Summary of API

There are three types of API in order to interact with PandoraPay DAEMON

1. HTTP
   1. [X] authentication
   2. [ ] notifications
2. HTTP RPC 
   1. [X] authentication
   2. [x] wallet
   3. [ ] notifications 
3. HTTP Websockets
   1. [X] authentication
   2. [x] wallet
   3. [X] notifications

## List of API 

List of all APIs

| REST API               | Description                                                                                                                                                                   | HTTP    | JSON RPC | HTTP Websocket | Explanation                                                       |
|------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|----------|----------------|-------------------------------------------------------------------|
| ping                   | Ping/Pong                                                                                                                                                                     | ✓ | ✓  | ✓        |                                                                   |
| "" (empty string)      | Node Info                                                                                                                                                                     | ✓ | ✓  | ✓        |                                                                   |
| chain                  | Blockchain summary                                                                                                                                                            | ✓ | ✓  | ✓        |                                                                   |
| blockchain             | alias for chain                                                                                                                                                               | ✓ | ✓  | ✓        |                                                                   |
| sync                   | Sync Info                                                                                                                                                                     | ✓ | ✓  | ✓        |                                                                   |
| block-hash             | Block hash from height                                                                                                                                                        | ✓ | ✓  | ✓        |                                                                   |
| block                  | Block with Txs hashes only                                                                                                                                                    | ✓ | ✓  | ✓        |                                                                   |
| block-complete         | Block with Txs                                                                                                                                                                | ✓ | ✓  | ✓        |                                                                   |
| block-miss-txs         | Block with Txs that are not specified in a transaction list                                                                                                                   | ✗ | ✗  | ✓        | Used only for Consensus                                           |
| tx-hash                | Tx hash from height                                                                                                                                                           | ✓ | ✓  | ✓        |                                                                   |
| tx                     | Transaction                                                                                                                                                                   | ✓ | ✓  | ✓        |                                                                   |
| account                | Account                                                                                                                                                                       | ✓ | ✓  | ✓        |                                                                   |
| accounts/count         | Number of accounts for an asset                                                                                                                                               | ✓ | ✓  | ✓        |                                                                   |
| accounts/keys-by-index | Accounts Keys for an asset specified by a list of indexes                                                                                                                     | ✓ | ✓  | ✓        |                                                                   |
| accounts/keys          | Accounts for an asset specified by a list of Accounts Keys                                                                                                                    | ✓ | ✓  | ✓        |                                                                   |
| asset                  | Asset                                                                                                                                                                         | ✓ | ✓  | ✓        |                                                                   |
| asset/fee-liquidity    | Asset Fee Liquidity                                                                                                                                                           | ✓ | ✓  | ✓        |                                                                   |
| mempool                | List of Tx Hashes that are in the mempool                                                                                                                                     | ✓ | ✓  | ✓        |                                                                   |
| mempool/tx-exists      | Existence of a Tx Hash in the mempool                                                                                                                                         | ✓ | ✓  | ✓        |                                                                   |
| mempool/new-tx         | Validate, Include and Broadcast Tx                                                                                                                                            | ✓ | ✓  | ✓        |                                                                   |
| mepool/new-tx-id       | Send a new txId to a node. In case the other node doesn't have this transaction in mempool, it will ask to download the transaction                                           | ✗ | ✗  | ✓        | websockets only*                                                  |
| network/nodes          | List of peers (50% of most active nodes, 50% of random nodes)                                                                                                                 | ✓ | ✓  | ✓        |                                                                   |
| asset-info             | Shorter version of an Asset                                                                                                                                                   | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| block-info             | Shorter version of a Block                                                                                                                                                    | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| tx-info                | Shorter version of a Tx                                                                                                                                                       | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| tx-preview             | Shorter version of a Tx                                                                                                                                                       | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| account/txs            | Account transactions                                                                                                                                                          | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| account/mempool        | Account pending transactions in mempool                                                                                                                                       | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| account/mempool-nonce  | Account new nonce from the mempool                                                                                                                                            | ✓ | ✓  | ✓        | Requires the node to be open with --seed-wallet-nodes-info="true" |
| handshake              | Websocket Handshake                                                                                                                                                           | ✗ | ✗  | ✓        | Used only in websockets                                           |
| chain-get              | Short information about Blockchain                                                                                                                                            | ✗ | ✗  | ✓        | Used only for Consensus                                           |
| chain-update           | Notify the node of a Blockchain Update                                                                                                                                        | ✗ | ✗  | ✓        | Used only for Consensus                                           |
| sub                    | Subscribe for changes in Account, PlainAccount, AccountTransactions, Asset, Registration and Transaction. The node will send a notification if the subscribed data is changed | ✗ | ✗  | ✓        | websockets only*                                                  |
| unsub                  | Unsubscribe from a change                                                                                                                                                     | ✗ | ✗  | ✓        | websockets only*                                                  |
| faucet/info            | Faucet information (hcaptcha)                                                                                                                                                 | ✓ | ✓  | ✓        | Requires the node to be open with --faucet-testnet-enabled="true" |
| faucet/coins           | Get Faucet coins                                                                                                                                                              | ✓ | ✓  | ✓        | Requires the node to be open with --faucet-testnet-enabled="true" |
| delegator-node/info    | Delegator Info                                                                                                                                                                | ✓ | ✓  | ✓        | Requires                                                          |
| delegator-node/ask     | Request                                                                                                                                                                       | ✓ | ✓  | ✓        | Requires                                                          |

TODO: TCP

## Integration to a third party app

The best and the most efficient way is to use the PaymentID attribute
and require the paymentIDs for each transaction. Instead of having a newly created address for 
every user or product/good, you should use a new PaymentID. By using this 
PaymentID, you can distinguish which user paid for which product/good was paid for. Your 
app should check all transactions, verify that something has 
really received and based on the paymentID to link and identify the user who paid for or the product/good that was paid for.

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.