# PandoraPay API

## Summary of API

There are three types of API in order to interact with PandoraPay DAEMON

1. HTTP
   1. [X] authentication
   2. [x] wallet
   3. [ ] notifications

   Data is packed using `json`

2. HTTP RPC 
   1. [X] authentication
   2. [x] wallet
   3. [ ] notifications
   
   Data is packed using `json`

3. HTTP Websockets
   1. [X] authentication
   2. [x] wallet
   3. [X] notifications

   Data is packed using `msgpack`

## List of API 

List of all APIs

| REST API               | Description                                                                                                                                                                   | HTTP | JSON RPC | HTTP Websocket | Requires Auth | Explanation                              |
|------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------|----------|----------------|---------------|------------------------------------------|
| ping                   | Ping/Pong                                                                                                                                                                     | ✓    | ✓        | ✓              |               |                                          |
| "" (empty string)      | Node Info                                                                                                                                                                     | ✓    | ✓        | ✓              |               |                                          |
| chain                  | Blockchain summary                                                                                                                                                            | ✓    | ✓        | ✓              |               |                                          |
| blockchain             | alias for chain                                                                                                                                                               | ✓    | ✓        | ✓              |               |                                          |
| sync                   | Sync Info                                                                                                                                                                     | ✓    | ✓        | ✓              |               |                                          |
| block-hash             | Block hash from height                                                                                                                                                        | ✓    | ✓        | ✓              |               |                                          |
| block                  | Block with Txs hashes only                                                                                                                                                    | ✓    | ✓        | ✓              |               |                                          |
| block-complete         | Block with Txs                                                                                                                                                                | ✓    | ✓        | ✓              |               |                                          |
| block-miss-txs         | Block with Txs that are not specified in a transaction list                                                                                                                   | ✗    | ✗        | ✓              |               | Used only for Consensus                  |
| tx-hash                | Tx hash from height                                                                                                                                                           | ✓    | ✓        | ✓              |               |                                          |
| tx                     | Transaction                                                                                                                                                                   | ✓    | ✓        | ✓              |               |                                          |
| account                | Account                                                                                                                                                                       | ✓    | ✓        | ✓              |               |                                          |
| accounts/count         | Number of accounts for an asset                                                                                                                                               | ✓    | ✓        | ✓              |               |                                          |
| accounts/keys-by-index | Accounts Keys for an asset specified by a list of indexes                                                                                                                     | ✓    | ✓        | ✓              |               |                                          |
| accounts/keys          | Accounts for an asset specified by a list of Accounts Keys                                                                                                                    | ✓    | ✓        | ✓              |               |                                          |
| asset                  | Asset                                                                                                                                                                         | ✓    | ✓        | ✓              |               |                                          |
| asset/fee-liquidity    | Asset Fee Liquidity                                                                                                                                                           | ✓    | ✓        | ✓              |               |                                          |
| mempool                | List of Tx Hashes that are in the mempool                                                                                                                                     | ✓    | ✓        | ✓              |               |                                          |
| mempool/tx-exists      | Existence of a Tx Hash in the mempool                                                                                                                                         | ✓    | ✓        | ✓              |               |                                          |
| mempool/new-tx         | Validate, Include and Broadcast Tx                                                                                                                                            | ✓    | ✓        | ✓              |               |                                          |
| mepool/new-tx-id       | Send a new txId to a node. In case the other node doesn't have this transaction in mempool, it will ask to download the transaction                                           | ✗    | ✗        | ✓              |               |                                          |
| network/nodes          | List of peers (50% of most active nodes, 50% of random nodes)                                                                                                                 | ✓    | ✓        | ✓              |               |                                          |
| asset-info             | Shorter version of an Asset                                                                                                                                                   | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| block-info             | Shorter version of a Block                                                                                                                                                    | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| tx-info                | Shorter version of a Tx                                                                                                                                                       | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| tx-preview             | Shorter version of a Tx                                                                                                                                                       | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| account/txs            | Account transactions                                                                                                                                                          | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| account/mempool        | Account pending transactions in mempool                                                                                                                                       | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| account/mempool-nonce  | Account new nonce from the mempool                                                                                                                                            | ✓    | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true" |
| handshake              | Websocket Handshake                                                                                                                                                           | ✗    | ✗        | ✓              |               | Used only in websockets                  |
| chain-get              | Short information about Blockchain                                                                                                                                            | ✗    | ✗        | ✓              |               | Used only for Consensus                  |
| chain-update           | Notify the node of a Blockchain Update                                                                                                                                        | ✗    | ✗        | ✓              |               | Used only for Consensus                  |
| sub                    | Subscribe for changes in Account, PlainAccount, AccountTransactions, Asset, Registration and Transaction. The node will send a notification if the subscribed data is changed | ✗    | ✗        | ✓              |               |                                          |
| unsub                  | Unsubscribe from a change                                                                                                                                                     | ✗    | ✗        | ✓              |               |                                          |
| faucet/info            | Faucet information (hcaptcha)                                                                                                                                                 | ✓    | ✓        | ✓              |               | Requires --faucet-testnet-enabled="true" |
| faucet/coins           | Get Faucet coins                                                                                                                                                              | ✓    | ✓        | ✓              |               | Requires --faucet-testnet-enabled="true" |
| delegator-node/info    | Delegator Info                                                                                                                                                                | ✓    | ✓        | ✓              |               | Requires                                 |
| delegator-node/ask     | Request                                                                                                                                                                       | ✓    | ✓        | ✓              |               | Requires                                 |
| login                  | Login user by providing credentials                                                                                                                                           | ✗    | ✗        | ✓              |               | Requires --auth-users                    |
| logout                 | Logout user from connection                                                                                                                                                   | ✗    | ✗        | ✓              | !             | Requires --auth-users                    |
| wallet/get-addresses   | Get all wallet accounts                                                                                                                                                       | ✓    | ✓        | ✓              | !             | Requires --auth-users                    |
| wallet/create-address  | Create a new empty address                                                                                                                                                    | ✓    | ✓        | ✓              | !             | Requires --auth-users                    |
| wallet/get-balances    | Get the balances (decrypted) of the requested wallet addresses                                                                                                                | ✓    | ✓        | ✓              | !             | Requires --auth-users                    |
| wallet/delete-address  | Delete an address from the wallet                                                                                                                                             | ✓    | ✓        | ✓              | !             | Requires --auth-users                    |


TODO: TCP

## Enable Authentication

To Set users and enable authentication use argument `--auth-users='[{"user": "username", "pass": "secret"}]'`

## Integration to a third party app

The best and the most efficient way is to use the PaymentID attribute
and require the paymentIDs for each transaction. Instead of having a newly created address for 
every user or product/good, you should use a new PaymentID. By using this 
PaymentID, you can distinguish which user paid for which product/good was paid for. Your 
app should check all transactions, verify that something has 
really received and based on the paymentID to link and identify the user who paid for or the product/good that was paid for.

## Examples of APIs

#### wallet/get-addresses
Request `curl http://127.0.0.1:5230/wallet/get-addresses?user=username&pass=password`

Output
```
{
    "version": 0,
    "encrypted": 0,
    "addresses": [ {
         "version": 0,
         "name": "Addr_0",
         "seedIndex": 0,
         "isMine": true,
         "privateKey": {
             "key": "82f9fa0ec4d13f39008ce2a8aab8169a6f1cf3a453b6f4ade19f36dcd675b175"
         },
         "registration": "16fb6f16f399dcd7dc1657444a033f80e9a7029e31fbaa4451f7b772c8703d7b0c117d768516a26707274ef595c084f7512582f00ee4d32359183b1a5cb3e8ce",
         "publicKey": "027140ac2fc222d87aee8dce2539b83aaa8882658cb23e9ebda18618361e5eb001",
         "balancesDecrypted": {
             "0000000000000000000000000000000000000000": {
                 "amount": 242927,
                 "asset": "0000000000000000000000000000000000000000"
             }
         },
         "addressEncoded": "PANDDEVAAJxQKwvwiLYeu6NziU5uDqqiIJljLI<nr2hhhg2Hl6wAQCT7qfa",
         "addressRegistrationEncoded": "PANDDEVAAJxQKwvwiLYeu6NziU5uDqqiIJljLI<nr2hhhg2Hl6wAQEKZR0gGenCYXf4jt<ZFx6<e7Nr0SIN9507FRvGT2jOIBGufNElj02S9aKZZ5G9FgmNN06oHjMgbiZRoYdW57NWvfkqfQ==",
         "delegatedStake": {
             "privateKey": {
                 "key": "3110c3e2c9bc8acb43bf930684adc96ffc83ed3795539ec106d8755464fe7b85"
             },
             "publicKey": "2cef23fc2b72689a1e1324a9da1e810972f42444f9b7f40a4fffe395f4d7190300",
             "lastKnownNonce": 0
         }
    }, ...
    ]
}
```

#### wallet/get-balances

Request Using PublicKey `curl http://127.0.0.1:5230/wallet/get-balances?list.0.publicKey=82f9fa0ec4d13f39008ce2a8aab8169a6f1cf3a453b6f4ade19f36dcd675b175&user=username&pass=password`

OR

Request Using Address `curl http://127.0.0.1:5230/wallet/get-balances?list.0.address=PANDDEVAAJxQKwvwiLYeu6NziU5uDqqiIJljLI<nr2hhhg2Hl6wAQCT7qfa&user=username&pass=password`

Output

```
{
    "results":[{
            "address": "PANDDEVABc3D9FePUuPADupO1p8jvtwEAVG5L3>sDttvmCw><jgAAABpTAR",
            "plainAcc": null,
            "balance":[ {
                    "amount": "15f8136864b1c06ebed9c03a006a61386d9d2c93310ff9758b7d3a5580a49a6d0018b822c42c27ad84d2971544d743417bb28ff7530f848e9d52c23598a665b01d01",
                    "value": 65205984,
                    "asset": "0000000000000000000000000000000000000000"
                }
            ]
        }
    ]
}
```

**Amount** is the encrypted balance using ElGamal

**Value** is the decrypted value.

WARNING! The decryting algorithm is a brute force. If you have more than 8 decimals values, it could take even a few minutes to decrypt the balance is case it was changed.

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.