package addresses

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/crypto/ecdsa"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key [32]byte
}

func (pk *PrivateKey) GeneratePublicKey() (publicKey [33]byte, err error) {

	var pub []byte
	if pub, err = ecdsa.ComputePublicKey(pk.Key[:]); err != nil {
		return
	}

	publicKey = *helpers.Byte33(pub)
	return
}

func (pk *PrivateKey) GenerateAddress(usePublicKeyHash bool, amount uint64, paymentID []byte) (*Address, error) {

	publicKey, err := ecdsa.ComputePublicKey(pk.Key[:])
	if err != nil {
		return nil, errors.New("Strange error. Your private key was invalid")
	}
	if len(paymentID) != 0 && len(paymentID) != 8 {
		return nil, errors.New("Your payment ID is invalid")
	}

	var finalPublicKey []byte

	var version AddressVersion

	if usePublicKeyHash {
		publicKeyHash := crypto.ComputePublicKeyHash(*helpers.Byte33(publicKey))
		finalPublicKey = publicKeyHash[:]
		version = SimplePublicKeyHash
	} else {
		finalPublicKey = publicKey
		version = SimplePublicKey
	}

	return &Address{
		config.NETWORK_SELECTED,
		version,
		finalPublicKey[:],
		amount,
		paymentID,
	}, nil
}

func (pk *PrivateKey) Sign(message *helpers.Hash) ([65]byte, error) {
	if len(message) != 32 {
		return [65]byte{}, errors.New("Message must be a hash")
	}
	privateKey, err := ecdsa.ToECDSA(pk.Key[:])
	if err != nil {
		return [65]byte{}, err
	}

	var signature []byte
	if signature, err = ecdsa.Sign(message[:], privateKey); err != nil {
		return [65]byte{}, err
	}
	return *helpers.Byte65(signature), nil
}

func GenerateNewPrivateKey() *PrivateKey {
	key := helpers.Byte32(helpers.RandomBytes(32))
	return &PrivateKey{Key: *key}
}
