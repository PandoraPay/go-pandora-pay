# Assets in PandoraPay

Transactions in PandoraPay can be of two types:
a. Simple Transaction (Public input)
b. Zether Transaction (Confidential amount and Ring members)

Transaction Scripts in PandoraPay

a. Simple Transactions
  1. **SCRIPT_UPDATE_DELEGATE** will update delegate information and/or convert unclaimed funds into staking. 
  2. **SCRIPT_UNSTAKE** will unstake a certain amount and move the amount to unclaimed funds.
  3. **SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY** will allow a liquidity offer for a certain asset. 
  
b. Zether Transaction
  1. **SCRIPT_TRANSFER** will transfer from an unknown sender to an unknown receiver an unknown amount. 
  4. **SCRIPT_ASSET_CREATE** will allow to create a new asset. The fee is paid by an unknown sender
  5. **SCRIPT_ASSET_SUPPLY_INCREASE** will allow to increase the supply of an asset X with value Y and move these to a known receiver address Z. The fee is paid by an unknown sender   

# DISCLAIMER:
This source code is released for research purposes only, with the intent of researching and studying a decentralized p2p network protocol.

PANDORAPAY IS AN OPEN SOURCE COMMUNITY DRIVEN RESEARCH PROJECT. THIS IS RESEARCH CODE PROVIDED TO YOU "AS IS" WITH NO WARRANTIES OF CORRECTNESS. IN NO EVENT SHALL THE CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES. USE AT YOUR OWN RISK.

You may not use this source code for any illegal or unethical purpose; including activities which would give rise to criminal or civil liability.

Under no event shall the Licensor be responsible for the activities, or any misdeeds, conducted by the Licensee.