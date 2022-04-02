package addresses

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
)

type PrivateKey struct {
	Key []byte `json:"key" msgpack:"key"` //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() []byte {
	return pk.Key[32:]
}

func (pk *PrivateKey) GeneratePublicKeyHash() []byte {
	pb := pk.Key[32:]
	return cryptography.GetPublicKeyHash(pb)
}

func (pk *PrivateKey) GenerateAddress(paymentID []byte, paymentAmount uint64, paymentAsset []byte) (*Address, error) {
	publicKeyHash := pk.GeneratePublicKeyHash()
	return NewAddr(config.NETWORK_SELECTED, SIMPLE_PUBLIC_KEY_HASH, publicKeyHash, paymentID, paymentAmount, paymentAsset)
}

//make sure message is a hash to avoid leaking any parts of the private key
func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	return ed25519.Sign(pk.Key, message), nil
}

func (pk *PrivateKey) Verify(message, signature []byte) bool {
	pb := pk.GeneratePublicKey()
	return ed25519.Verify(pb, message, signature)
}

func (pk *PrivateKey) Decrypt(message []byte) ([]byte, error) {
	return nil, errors.New("Encryption is not supported right now")
}

func NewPrivateKey(key []byte) (*PrivateKey, error) {
	if len(key) != cryptography.PrivateKeySize {
		return nil, errors.New("Private Key length is invalid")
	}
	return &PrivateKey{Key: key}, nil
}

func GenerateNewPrivateKey() *PrivateKey {
	var err error
	var privateKey []byte
	for _, privateKey, err = ed25519.GenerateKey(rand.Reader); err != nil; {
		continue
	}
	return &PrivateKey{Key: privateKey}
}

func CreatePrivateKeyFromSeed(key []byte) (*PrivateKey, error) {
	if len(key) != cryptography.PrivateKeySize {
		return nil, errors.New("Private key length is invalid")
	}
	return &PrivateKey{Key: key}, nil
}
