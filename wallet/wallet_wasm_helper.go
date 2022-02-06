package wallet

import (
	"encoding/hex"
	"pandora-pay/helpers"
)

func (wallet *Wallet) GetDataForDecodingBalance(publicKey, asset []byte) (privateKey helpers.HexBytes, previousValue uint64) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]
	privateKey = addr.PrivateKey.Key

	if addr.BalancesDecoded[hex.EncodeToString(asset)] != nil {
		previousValue = addr.BalancesDecoded[hex.EncodeToString(asset)].AmountDecoded
	}

	return
}
