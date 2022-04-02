package addresses

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
)

func TestGenerateNewPrivateKey(t *testing.T) {

	privateKey := GenerateNewPrivateKey()
	assert.Equal(t, len(privateKey.Key), cryptography.PrivateKeySize, "Invalid private key length")
	assert.Equal(t, bytes.Equal(privateKey.Key, helpers.EmptyBytes(cryptography.PrivateKeySize)), false, "Invalid private key is empty")
	assert.Equal(t, bytes.Equal(privateKey.Key, privateKey.GeneratePublicKey()), false)

}

func TestPrivateKey_GenerateAddress(t *testing.T) {

	privateKey := GenerateNewPrivateKey()

	address, err := privateKey.GenerateAddress(nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, bytes.Equal(address.PublicKeyHash, privateKey.GeneratePublicKeyHash()), true)
	assert.Equal(t, bytes.Equal(address.PublicKeyHash, privateKey.Key), false)

	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.Equal(t, false, bytes.Equal(address.PublicKeyHash, helpers.EmptyBytes(cryptography.PublicKeySize)))
	assert.Equal(t, address.PaymentAmount, uint64(0))
	assert.Equal(t, len(address.PaymentID), 0)

	address, err = privateKey.GenerateAddress(nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.Equal(t, bytes.Equal(address.PublicKeyHash, helpers.EmptyBytes(cryptography.PublicKeyHashSize)), false)
	assert.Equal(t, address.PaymentAmount, uint64(0))
	assert.Equal(t, len(address.PaymentID), 0)

	address, err = privateKey.GenerateAddress(helpers.RandomBytes(8), 20, nil)
	assert.NoError(t, err)

	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.Equal(t, bytes.Equal(address.PublicKeyHash, helpers.EmptyBytes(cryptography.PublicKeyHashSize)), false)
	assert.Equal(t, address.PaymentAmount, uint64(20))
	assert.Equal(t, len(address.PaymentID), 8)

}
