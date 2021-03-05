package addresses

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key [32]byte
}

func (pk *PrivateKey) GeneratePublicKeySilent() (publicKey [33]byte, err error) {

	defer func() {
		if err2 := recover(); err2 != nil {
			err = helpers.ConvertRecoverError(err2)
		}
	}()
	publicKey = pk.GeneratePublicKey()
	return
}

func (pk *PrivateKey) GeneratePublicKey() [33]byte {

	pub, err := ecdsa.ComputePublicKey(pk.Key[:])
	if err != nil {
		panic(err)
	}

	return *helpers.Byte33(pub)
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
		publicKeyHash := cryptography.ComputePublicKeyHash(*helpers.Byte33(publicKey))
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

func (pk *PrivateKey) SignSilent(message helpers.Hash) (out [65]byte, err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = helpers.ConvertRecoverError(err2)
		}
	}()
	out = pk.Sign(message)
	return
}

func (pk *PrivateKey) Sign(message helpers.Hash) [65]byte {

	privateKey, err := ecdsa.ToECDSA(pk.Key[:])
	if err != nil {
		panic(err)
	}

	var signature []byte
	if signature, err = ecdsa.Sign(message[:], privateKey); err != nil {
		panic(err)
	}
	return *helpers.Byte65(signature)
}

func GenerateNewPrivateKey() *PrivateKey {
	key := helpers.Byte32(helpers.RandomBytes(32))
	return &PrivateKey{Key: *key}
}
