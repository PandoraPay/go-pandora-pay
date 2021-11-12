# Assets in PandoraPay

Private Assets can be created and transferred privately by anyone.

The following options can be done with assets:
1. Create Asset
2. Increase Supply
3. Reduce Supply
4. Transfer
5. Upgrade

## Create Asset

To create an asset you need to use the CLI command: "Private Asset Create". The creation is also made privately by using Zether private transactions.

Follow the command. For Asset JSON use the following template. The JSON needs to be oneliner.

```
{"name": "My Asset", "ticker": "AST", "description": "My simple Asset", "version": 0, "canUpgrade": true, "canMint": true, "canBurn": true, "canChangeUpdatePublicKey": true, "canChangeSupplyPublicKey": true, "canPause": false, "canFreeze": false, "decimalSeparator": 5, "maxSupply": 21000000000000, "supply": 0, "updatePublicKey": "", "supplyPublicKey": ""}
```

In case `updatePublicKey` or `supplyPublicKey` is not supplied the daemon will create you new key pairs and will store them in a separate file.

## Increase Supply

## Decrease Supply

## Transfer

Assets can be transfered using "Private Transfer" or in the web wallet.