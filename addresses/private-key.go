package addresses

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key helpers.HexBytes //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() ([]byte, error) {
	return ecdsa.ComputePublicKey(pk.Key)
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
		address.Version = SimplePublicKeyHash
	} else {
		address.Version = SimplePublicKey
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
