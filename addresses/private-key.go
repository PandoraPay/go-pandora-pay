package addresses

import (
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key helpers.HexBytes `json:"key"` //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() ([]byte, error) {
	return ecdsa.ComputePublicKey(pk.Key)
}

func (pk *PrivateKey) GeneratePublicKeyHash() ([]byte, error) {
	publicKey, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		return nil, err
	}
	return cryptography.ComputePublicKeyHash(publicKey), nil
}

func (pk *PrivateKey) GeneratePairs() ([]byte, []byte, error) {
	publicKey, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		return nil, nil, err
	}
	publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)
	return publicKey, publicKeyHash, nil
}

func (pk *PrivateKey) GenerateAddress(usePublicKeyHash bool, amount uint64, paymentID []byte) (*Address, error) {

	var publicKey, publicKeyHash []byte
	var version AddressVersion

	publicKey, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		return nil, err
	}

	if usePublicKeyHash {
		publicKeyHash = cryptography.ComputePublicKeyHash(publicKey)
		publicKey = []byte{}
		version = SIMPLE_PUBLIC_KEY_HASH
	} else {
		version = SIMPLE_PUBLIC_KEY
	}

	return NewAddr(config.NETWORK_SELECTED, version, publicKey, publicKeyHash, amount, paymentID)
}

func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	privateKey, err := ecdsa.ToECDSA(pk.Key)
	if err != nil {
		return nil, err
	}

	return ecdsa.Sign(message, privateKey)
}

func GenerateNewPrivateKey() *PrivateKey {
	return &PrivateKey{Key: helpers.RandomBytes(32)}
}
