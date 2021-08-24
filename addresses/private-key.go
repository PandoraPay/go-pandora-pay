package addresses

import (
	"pandora-pay/config"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/cryptography/ecies"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key helpers.HexBytes `json:"key"` //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() ([]byte, error) {
	return ecdsa.ComputePublicKey(pk.Key)
}

func (pk *PrivateKey) GeneratePairs() ([]byte, error) {
	publicKey, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func (pk *PrivateKey) GenerateAddress(amount uint64, paymentID []byte) (*Address, error) {

	var publicKey []byte
	var version AddressVersion

	publicKey, err := ecdsa.ComputePublicKey(pk.Key)
	if err != nil {
		return nil, err
	}

	version = SIMPLE_PUBLIC_KEY

	return NewAddr(config.NETWORK_SELECTED, version, publicKey, amount, paymentID)
}

func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	privateKey, err := ecdsa.ToECDSA(pk.Key)
	if err != nil {
		return nil, err
	}

	return ecdsa.Sign(message, privateKey)
}

func (pk *PrivateKey) Decrypt(message []byte) ([]byte, error) {
	privateKey, err := ecdsa.ToECDSA(pk.Key)
	if err != nil {
		return nil, err
	}

	priv := ecies.ImportECDSA(privateKey)
	out, err := priv.Decrypt(message, nil, nil)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func GenerateNewPrivateKey() *PrivateKey {
	return &PrivateKey{Key: helpers.RandomBytes(32)}
}
