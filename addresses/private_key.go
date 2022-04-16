package addresses

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	KeyWIF
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
	return CreateAddr(publicKeyHash, paymentID, paymentAmount, paymentAsset)
}

//make sure message is a hash to avoid leaking any parts of the private key
func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	return cryptography.SignMessage(pk.Key, message), nil
}

func (pk *PrivateKey) Verify(message, signature []byte) bool {
	return cryptography.VerifySignature(pk.GeneratePublicKey(), message, signature)
}

func (pk *PrivateKey) Decrypt(message []byte) ([]byte, error) {
	return nil, errors.New("Encryption is not supported right now")
}

func (pk *PrivateKey) Deserialize(buffer []byte) error {
	return pk.deserialize(buffer, cryptography.PrivateKeySize)
}

func GenerateNewPrivateKey() *PrivateKey {
	for {

		privateKey, err := NewPrivateKey(helpers.RandomBytes(cryptography.PrivateKeySize))
		if err != nil {
			continue
		}
		return privateKey
	}
}

func NewPrivateKey(key []byte) (*PrivateKey, error) {

	if len(key) != cryptography.PrivateKeySize {
		return nil, errors.New("Private Key length is invalid")
	}

	privateKey := &PrivateKey{
		KeyWIF{
			SIMPLE_PRIVATE_KEY_WIF,
			config.NETWORK_SELECTED,
			key,
			nil,
		},
	}

	privateKey.Checksum = privateKey.computeCheckSum()

	return privateKey, nil
}
