package addresses

import (
	"errors"
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

func (pk *PrivateKey) GeneratePairs() (publicKey []byte, publicKeyHash []byte, err error) {
	if publicKey, err = ecdsa.ComputePublicKey(pk.Key); err != nil {
		return
	}
	publicKeyHash = cryptography.ComputePublicKeyHash(publicKey)
	return
}

func (pk *PrivateKey) GenerateAddress(usePublicKeyHash bool, amount uint64, paymentID []byte) (address *Address, err error) {

	address = &Address{
		Network:   config.NETWORK_SELECTED,
		Amount:    amount,
		PaymentID: paymentID,
	}

	if address.PublicKey, err = ecdsa.ComputePublicKey(pk.Key); err != nil {
		return
	}
	if len(paymentID) != 0 && len(paymentID) != 8 {
		return nil, errors.New("Your payment ID is invalid")
	}

	address.PublicKeyHash = cryptography.ComputePublicKeyHash(address.PublicKey)

	if usePublicKeyHash {
		address.PublicKey = []byte{}
		address.Version = SIMPLE_PUBLIC_KEY_HASH
	} else {
		address.Version = SIMPLE_PUBLIC_KEY
	}

	return
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
