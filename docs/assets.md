# Assets in PandoraPay

Private Assets can be created and transferred privately by anyone.

The following options can be done with assets:
1. Create Asset
2. Increase Supply
3. Reduce Supply
4. Transfer
5. Upgrade

## Create Asset

To create an asset you need to use the CLI command: "Private Asset Create". 
The creation is also made privately by using Zether private transactions.

Follow the command. For Asset JSON use the following template. The JSON needs to be oneliner.

```
{"name": "My Asset", "ticker": "AST", "description": "My simple Asset", "version": 0, "canUpgrade": true, "canMint": true, "canBurn": true, "canChangeUpdatePublicKey": true, "canChangeSupplyPublicKey": true, "canPause": false, "canFreeze": false, "decimalSeparator": 5, "maxSupply": 21000000000000, "supply": 0, "updatePublicKey": "", "supplyPublicKey": ""}
```

In case `updatePublicKey` or `supplyPublicKey` is not supplied, the daemon will create you new key pairs and will store them in a separate file.

## Increase Supply

## Decrease Supply

## Transfer

Assets can be transferred using "Private Transfer" or in the web wallet.


# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.