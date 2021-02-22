package addresses

import (
	"math/rand"
	"pandora-pay/blockchain"
	"pandora-pay/crypto"
	"pandora-pay/gui"
)

type PrivateKey struct {
	Key []byte
}

func (pk *PrivateKey) GenerateTransparentAddress(usePublicKeyHash bool, amount uint64, paymentID *[]byte) *Address {

	publicKey, err := crypto.GeneratePublicKey(pk.Key)
	if err != nil {
		gui.Error("Strange error. Your private key was invalid", err)
		return nil
	}

	var finalPublicKey []byte

	if usePublicKeyHash {
		finalPublicKey = crypto.RIPEMD(publicKey)
	} else {
		finalPublicKey = publicKey
	}

	return &Address{Network: blockchain.NETWORK_SELECTED, Version: uint64(TransparentPublicKey), PublicKey: finalPublicKey[:], Amount: amount, PaymentID: *paymentID}
}

func GenerateNewPrivateKey() *PrivateKey {

	var key []byte
	key = make([]byte, 32)
	rand.Read(key)

	return &PrivateKey{Key: key}
}
