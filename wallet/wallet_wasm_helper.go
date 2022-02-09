package wallet

import (
	"encoding/hex"
	"pandora-pay/helpers"
)

func (wallet *Wallet) GetDataForDecryptingBalance(publicKey, asset []byte) (privateKey helpers.HexBytes, previousValue uint64) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]
	privateKey = addr.PrivateKey.Key

	if addr.BalancesDecrypted[hex.EncodeToString(asset)] != nil {
		previousValue = addr.BalancesDecrypted[hex.EncodeToString(asset)].AmountDecrypted
	}

	return
}
