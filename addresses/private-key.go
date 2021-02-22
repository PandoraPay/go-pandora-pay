package addresses

import (
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key []byte
}

func (pk *PrivateKey) GenerateTransparentAddress(usePublicKeyHash bool, amount uint64, paymentID []byte) (*Address, error) {

	publicKey, err := crypto.GeneratePublicKey(pk.Key)
	if err != nil {
		return nil, errors.New("Strange error. Your private key was invalid")
	}
	if len(paymentID) != 0 && len(paymentID) != 8 {
		return nil, errors.New("Your payment ID is invalid")
	}

	var finalPublicKey []byte

	var version AddressVersion

	if usePublicKeyHash {
		finalPublicKey = crypto.RIPEMD(publicKey)
		version = AddressVersionTransparentPublicKeyHash
	} else {
		finalPublicKey = publicKey
		version = AddressVersionTransparentPublicKey
	}

	return &Address{Network: blockchain.NETWORK_SELECTED, Version: version, PublicKey: finalPublicKey[:], Amount: amount, PaymentID: paymentID}, nil
}

func GenerateNewPrivateKey() *PrivateKey {

	return &PrivateKey{Key: helpers.RandomBytes(32)}
}
