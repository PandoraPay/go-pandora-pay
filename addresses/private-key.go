package addresses

import (
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key []byte //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() []byte {

	pub, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		panic(err)
	}

	return pub
}

func (pk *PrivateKey) GenerateAddress(usePublicKeyHash bool, amount uint64, paymentID []byte) *Address {

	publicKey, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		panic("Strange error. Your private key was invalid")
	}
	if len(paymentID) != 0 && len(paymentID) != 8 {
		panic("Your payment ID is invalid")
	}

	publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)

	var version AddressVersion

	if usePublicKeyHash {
		publicKey = []byte{}
		version = SimplePublicKeyHash
	} else {
		version = SimplePublicKey
	}

	return &Address{
		config.NETWORK_SELECTED,
		version,
		publicKey,
		publicKeyHash,
		amount,
		paymentID,
	}
}

func (pk *PrivateKey) Sign(message []byte) []byte {
	privateKey, err := ecdsa.ToECDSA(pk.Key)
	if err != nil {
		panic(err)
	}

	var signature []byte
	if signature, err = ecdsa.Sign(message, privateKey); err != nil {
		panic(err)
	}
	return signature
}

func GenerateNewPrivateKey() *PrivateKey {
	return &PrivateKey{Key: helpers.RandomBytes(32)}
}
