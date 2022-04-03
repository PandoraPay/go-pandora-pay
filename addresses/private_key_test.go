package addresses

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"testing"
)

func TestGenerateNewPrivateKey(t *testing.T) {

	privateKey := GenerateNewPrivateKey()
	assert.Equal(t, len(privateKey.Key), cryptography.PrivateKeySize, "Invalid private key length")
	assert.NotEqual(t, privateKey.Key, helpers.EmptyBytes(cryptography.PrivateKeySize), "Invalid private key is empty")
	assert.NotEqual(t, privateKey.Key, privateKey.GeneratePublicKey())

}

func TestPrivateKey_GenerateAddress(t *testing.T) {

	privateKey := GenerateNewPrivateKey()

	address, err := privateKey.GenerateAddress(false, nil, false, nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, address.PublicKey, privateKey.GeneratePublicKey())
	assert.NotEqual(t, address.PublicKey, privateKey.Key)

	assert.Equal(t, len(address.PublicKey), cryptography.PublicKeySize)
	assert.NotEqual(t, address.PublicKey, helpers.EmptyBytes(cryptography.PublicKeySize))
	assert.Equal(t, address.PaymentAmount, uint64(0))
	assert.Equal(t, len(address.PaymentID), 0)

	address, err = privateKey.GenerateAddress(false, nil, false, nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, len(address.PublicKey), cryptography.PublicKeySize)
	assert.NotEqual(t, address.PublicKey, helpers.EmptyBytes(cryptography.PublicKeySize))
	assert.Equal(t, address.PaymentAmount, uint64(0))
	assert.Equal(t, len(address.PaymentID), 0)

	address, err = privateKey.GenerateAddress(false, nil, false, helpers.RandomBytes(8), 20, nil)
	assert.NoError(t, err)

	assert.Equal(t, len(address.PublicKey), cryptography.PublicKeySize)
	assert.NotEqual(t, address.PublicKey, helpers.EmptyBytes(cryptography.PublicKeySize))
	assert.Equal(t, address.PaymentAmount, uint64(20))
	assert.Equal(t, len(address.PaymentID), 8)

}

func TestPrivateKey_BN256(t *testing.T) {
	privateKey := GenerateNewPrivateKey()
	address, err := privateKey.GenerateAddress(false, nil, false, nil, 0, helpers.EmptyBytes(0))
	assert.NoError(t, err)

	priv := new(crypto.BNRed).SetBytes(privateKey.Key)
	pub := crypto.GPoint.ScalarMult(priv)
	pubKey := pub.EncodeCompressed()

	assert.Equal(t, address.PublicKey, pubKey)
}
