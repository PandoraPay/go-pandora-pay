# PandoraPay API

## Summary of API

There are three types of API in order to interact with PandoraPay DAEMON

Binary data is passed using **base64** not HEX! Passing GET arguments requires **URLEncode**.

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

| REST API                | Description                                                                                                                                                                   | HTTP GET | HTTP POST | JSON RPC | HTTP Websocket | Requires Auth | Explanation                                                                                                                                                                                                                                                                                                                                                                                     |
|-------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------|-----------|----------|----------------|---------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ping                    | Ping/Pong                                                                                                                                                                     | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| "" (empty string)       | Node Info                                                                                                                                                                     | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| chain                   | Blockchain summary                                                                                                                                                            | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| blockchain              | alias for chain                                                                                                                                                               | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| sync                    | Sync Info                                                                                                                                                                     | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| block-hash              | Block hash from height                                                                                                                                                        | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| block                   | Block with Txs hashes only                                                                                                                                                    | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| block-complete          | Block with Txs                                                                                                                                                                | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| block-miss-txs          | Block with Txs that are not specified in a transaction list                                                                                                                   | ✗        | ✗         | ✗        | ✓              |               | Used only for Consensus                                                                                                                                                                                                                                                                                                                                                                         |
| tx-hash                 | Tx hash from height                                                                                                                                                           | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| tx                      | Transaction                                                                                                                                                                   | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| tx-raw                  | Transaction serialized                                                                                                                                                        | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| account                 | Account                                                                                                                                                                       | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| accounts/count          | Number of accounts for an asset                                                                                                                                               | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| accounts/keys-by-index  | Accounts Keys for an asset specified by a list of indexes                                                                                                                     | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| accounts/keys           | Accounts for an asset specified by a list of Accounts Keys                                                                                                                    | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| asset                   | Asset                                                                                                                                                                         | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| asset/fee-liquidity     | Asset Fee Liquidity                                                                                                                                                           | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| mempool                 | List of Tx Hashes that are in the mempool                                                                                                                                     | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| mempool/tx-exists       | Existence of a Tx Hash in the mempool                                                                                                                                         | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| mempool/new-tx          | Validate, Include and Broadcast Tx                                                                                                                                            | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| mepool/new-tx-id        | Send a new txId to a node. In case the other node doesn't have this transaction in mempool, it will ask to download the transaction                                           | ✗        | ✗         | ✗        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| network/nodes           | List of peers (50% of most active nodes, 50% of random nodes)                                                                                                                 | ✓        | ✗         | ✓        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| asset-info              | Shorter version of an Asset                                                                                                                                                   | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| block-info              | Shorter version of a Block                                                                                                                                                    | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| tx-info                 | Shorter version of a Tx                                                                                                                                                       | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| tx-preview              | Shorter version of a Tx                                                                                                                                                       | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| account/txs             | Account transactions                                                                                                                                                          | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| account/mempool         | Account pending transactions in mempool                                                                                                                                       | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| account/mempool-nonce   | Account new nonce from the mempool                                                                                                                                            | ✓        | ✗         | ✓        | ✓              |               | Requires --seed-wallet-nodes-info="true"                                                                                                                                                                                                                                                                                                                                                        |
| handshake               | Websocket Handshake                                                                                                                                                           | ✗        | ✗         | ✗        | ✓              |               | Used only in websockets                                                                                                                                                                                                                                                                                                                                                                         |
| get-chain               | Short information about Blockchain                                                                                                                                            | ✗        | ✗         | ✗        | ✓              |               | Used only for Consensus                                                                                                                                                                                                                                                                                                                                                                         |
| chain-update            | Notify the node of a Blockchain Update                                                                                                                                        | ✗        | ✗         | ✗        | ✓              |               | Used only for Consensus                                                                                                                                                                                                                                                                                                                                                                         |
| sub                     | Subscribe for changes in Account, PlainAccount, AccountTransactions, Asset, Registration and Transaction. The node will send a notification if the subscribed data is changed | ✗        | ✗         | ✗        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| unsub                   | Unsubscribe from a change                                                                                                                                                     | ✗        | ✗         | ✗        | ✓              |               |                                                                                                                                                                                                                                                                                                                                                                                                 |
| faucet/info             | Faucet information (hcaptcha)                                                                                                                                                 | ✓        | ✗         | ✓        | ✓              |               | Requires --faucet-testnet-enabled="true"                                                                                                                                                                                                                                                                                                                                                        |
| faucet/coins            | Get Faucet coins                                                                                                                                                              | ✓        | ✗         | ✓        | ✓              |               | Requires --faucet-testnet-enabled="true"                                                                                                                                                                                                                                                                                                                                                        |
| delegator-node/info     | Delegator Info                                                                                                                                                                | ✓        | ✗         | ✓        | ✓              |               | Requires                                                                                                                                                                                                                                                                                                                                                                                        |
| delegator-node/ask      | Request                                                                                                                                                                       | ✓        | ✗         | ✓        | ✓              |               | Requires                                                                                                                                                                                                                                                                                                                                                                                        |
| login                   | Login user by providing credentials                                                                                                                                           | ✗        | ✗         | ✗        | ✓              |               | Requires --auth-users                                                                                                                                                                                                                                                                                                                                                                           |
| logout                  | Logout user from connection                                                                                                                                                   | ✗        | ✗         | ✗        | ✓              | !             | Requires --auth-users                                                                                                                                                                                                                                                                                                                                                                           |
| wallet/get-addresses    | Get all wallet accounts                                                                                                                                                       | ✓        | ✗         | ✓        | ✓              | !             | Requires --auth-users                                                                                                                                                                                                                                                                                                                                                                           |
| wallet/create-address   | Create a new empty address                                                                                                                                                    | ✓        | ✗         | ✓        | ✓              | !             | Requires --auth-users                                                                                                                                                                                                                                                                                                                                                                           |
| wallet/get-balances     | Get the balances (decrypted) of the requested wallet addresses                                                                                                                | ✓        | ✗         | ✓        | ✓              | !             | It will load the balances and decrypt them. The decryption is a brute force algorithm that will check all balances until is found. Having an 8 decimal balance will take a few minutes! Requires --auth-users.                                                                                                                                                                                  |
| wallet/delete-address   | Delete an address from the wallet                                                                                                                                             | ✓        | ✗         | ✓        | ✓              | !             | Requires --auth-users                                                                                                                                                                                                                                                                                                                                                                           |
| wallet/decrypt-tx       | Decrypt a transaction using wallet                                                                                                                                            | ✓        | ✗         | ✓        | ✓              | !             | Will decrypt zether transaction and return Recipient Ring Position (if you are the sender), shared decrypted message and decrypted amount using Whisper protocol. The decrypted tx amount is checked fast by verifying only that the whisper amounts are indeed the real values. In case the whisper amount is wrong, the call will return false and report the amount 0. Requires --auth-users |
| wallet/private-transfer | Create a private Transfer                                                                                                                                                     | ✗        | ✓         | ✓        | ✓              | !             | It will create and broadcast a private transaction. Requires --auth-users                                                                                                                                                                                                                                                                                                                       |



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

### wallet/get-addresses
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
             "key": "gvn6DsTRPzkAjOKoqrgWmm8c86RTtvSt4Z823NZ1sXU="
         },
         "registration": "FvtvFvOZ3NfcFldESgM/gOmnAp4x+6pEUfe3cshwPXsMEX12hRaiZwcnTvWVwIT3USWC8A7k0yNZGDsaXLPozg==",
         "publicKey": "AnFArC/CIth67o3OJTm4OqqIgmWMsj6evaGGGDYeXrAB",
         "decryptedBalances": {
             "AAAAAAAAAAAAAAAAAAAAAAAAAAA=": {
                 "amount": 242927,                
                 "encryptedBalance": "FfgTaGSxwG6+2cA6AGphOG2dLJMxD/l1i306VYCkmm0AGLgixCwnrYTSlxVE10NBe7KP91MPhI6dUsI1mKZlsB0B",
             }
         },
         "addressEncoded": "PANDDEVAAJxQKwvwiLYeu6NziU5uDqqiIJljLI<nr2hhhg2Hl6wAQCT7qfa",
         "addressRegistrationEncoded": "PANDDEVAAJxQKwvwiLYeu6NziU5uDqqiIJljLI<nr2hhhg2Hl6wAQEKZR0gGenCYXf4jt<ZFx6<e7Nr0SIN9507FRvGT2jOIBGufNElj02S9aKZZ5G9FgmNN06oHjMgbiZRoYdW57NWvfkqfQ==",
         "delegatedStake": {
             "privateKey": {
                 "key": "MRDD4sm8istDv5MGhK3Jb/yD7TeVU57BBth1VGT+e4U="
             },
             "publicKey": "LO8j/CtyaJoeEySp2h6BCXL0JET5t/QKT//jlfTXGQMA",
             "lastKnownNonce": 0
         }
    }, ...
    ]
}
```

### wallet/get-balances

Request Using PublicKey `curl http://127.0.0.1:5230/wallet/get-balances?list.0.publicKey=EkgfeoxQYNAeDTR%2BXz85AG8mHEhsPYM8fFSslBsgO7EB&user=username&pass=password`

OR

Request Using Address `curl http://127.0.0.1:5230/wallet/get-balances?list.0.address=PANDDEVAAJxQKwvwiLYeu6NziU5uDqqiIJljLI<nr2hhhg2Hl6wAQCT7qfa&user=username&pass=password`

Output

```
{
    "results":[{
            "address": "PANDDEVABc3D9FePUuPADupO1p8jvtwEAVG5L3>sDttvmCw><jgAAABpTAR",
            "plainAcc": null,
            "balances":[ {
                    "balance": "FfgTaGSxwG6+2cA6AGphOG2dLJMxD/l1i306VYCkmm0AGLgixCwnrYTSlxVE10NBe7KP91MPhI6dUsI1mKZlsB0B",
                    "amount": 65205984,
                    "asset": "AAAAAAAAAAAAAAAAAAAAAAAAAAA="
                }
            ]
        }
    ]
}
```

**balance** is the ElGamal encrypted balance 

**amount** is the decrypted value.

**WARNING!** The decryptor is a making brute force trying all possible balances starting from 0. If you have more than 8 decimals values, it could take even a few minutes to decrypt the balance is case it was changed.

### wallet/decrypt-tx

Request Using TxHash `curl http://127.0.0.1:5230/wallet/decrypt-tx?hash=dKTfcDJ4gRcV1Rx5ZFtXxsrh2YwlaljDLast5g3f1rY%3D&user=username&pass=password`

Output
```
{
   "decrypted":{
      "type":1,
      "zetherTx":{
         "payloads":[ {
               "whisperSenderValid":true,
               "sentAmount":100703740,
               "whisperRecipientValid":false,
               "receivedAmount":0,
               "recipientIndex":9,
               "message":"VGVzdG5ldCBGYXVjZXQgVHg="
            }
         ]
      }
   }
}
```

**whisperSenderValid** true if you were the sender and the whisper encrypted amount was successfully verified. 

**sentAmount**  amount if you were the sender and the whisper encrypted amount was successfully verified.

**recipientIndex** ring member position of the recipient if you were the sender

**whisperRecipientValid** true if you were the recipient and the whisper encrypted amount was successfully verified. 

**receivedAmount** amount if you were the recipient and the whisper encrypted amount was successfully verified.

**message** decrypted shared messaged

In case the whisper is malformed it will return accordingly.

### wallet/private-transfer

Creating private transfer using a POST request like the following:
```
curl -X POST  \
-H 'Content-Type: application/json'  \
-d '{ "user": "username", "pass": "password", "data": { "payloads": [ {"sender":  "PANDDEVAAaBVqiVyecV\u003cysBwcT\u003cGRkIHPBdbHZ9hwaS4wfV4xKYAQAPLjdy",  "recipient":  "PANDDEVABjp7xeB<oGlMe5PdvIq7oGhUq3iquvERZS3<Ax6CCzqAABnVMdN",  "amount": 100 }] }, "propagate": true }' http://127.0.0.1:5232/wallet/private-transfer
```

**WARNING!** When creating a private transfer, the balance must be decrypted for signing. The decryptor is a making brute force trying all possible balances starting from 0. If you have more than 8 decimals values, it could take even a few minutes to decrypt the balance is case it was changed.

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.