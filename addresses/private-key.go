package addresses

import (
	"pandora-pay/config"
	"pandora-pay/cryptography/cryptolib"
	"pandora-pay/helpers"
)

type PrivateKey struct {
	Key helpers.HexBytes `json:"key"` //32 byte
}

func (pk *PrivateKey) GeneratePublicKey() []byte {
	priv := new(cryptolib.BNRed).SetBytes(pk.Key)
	publicKey := cryptolib.GPoint.ScalarMult(priv)
	return publicKey.EncodeCompressed()
}

func (pk *PrivateKey) GenerateAddress(amount uint64, paymentID []byte) (*Address, error) {
	publicKey := pk.GeneratePublicKey()
	return NewAddr(config.NETWORK_SELECTED, SIMPLE_PUBLIC_KEY, publicKey, amount, paymentID)
}

//make sure message is a hash to avoid leaking any parts of the private key
func (pk *PrivateKey) Sign(message []byte) ([]byte, error) {
	return cryptolib.SignMessage(message, pk.Key)
}

func (pk *PrivateKey) Decrypt(message []byte) ([]byte, error) {
	panic("not implemented")
}

func GenerateNewPrivateKey() *PrivateKey {
	seed := cryptolib.RandomScalarBNRed()
	privateKey := seed.ToBytes()

	return &PrivateKey{Key: privateKey}
}

func CreatePrivateKeyFromSeed(seed []byte) *PrivateKey {
	return &PrivateKey{Key: seed}
}
