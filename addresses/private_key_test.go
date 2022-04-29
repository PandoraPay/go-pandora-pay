package addresses

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/cryptography"
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

	address, err := privateKey.GenerateAddress(nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, address.PublicKeyHash, privateKey.GeneratePublicKeyHash())
	assert.NotEqual(t, address.PublicKeyHash, privateKey.Key)

	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.NotEqual(t, address.PublicKeyHash, helpers.EmptyBytes(cryptography.PublicKeyHashSize))
	assert.Equal(t, address.PaymentAmount, uint64(0))
	assert.Equal(t, len(address.PaymentID), 0)

	address, err = privateKey.GenerateAddress(nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.NotEqual(t, address.PublicKeyHash, helpers.EmptyBytes(cryptography.PublicKeyHashSize))
	assert.Equal(t, address.PaymentAmount, uint64(0))
	assert.Equal(t, len(address.PaymentID), 0)

	address, err = privateKey.GenerateAddress(helpers.RandomBytes(8), 20, nil)
	assert.NoError(t, err)

	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.NotEqual(t, address.PublicKeyHash, helpers.EmptyBytes(cryptography.PublicKeyHashSize))
	assert.Equal(t, address.PaymentAmount, uint64(20))
	assert.Equal(t, len(address.PaymentID), 8)

}

func TestPrivateKey_GenerateAddressWEBD(t *testing.T) {

	privateKey, err := NewPrivateKey(helpers.DecodeHex("8e2f76671325e48c19b134fc62e8c5957a5bcb76cf37fb590c8069857122a7a7eb00511be600f3eddb9a8c13d0223001fcb72c37a005a96f1c77e8d35a8bc434"))
	assert.NoError(t, err)

	assert.Equal(t, helpers.DecodeHex("eb00511be600f3eddb9a8c13d0223001fcb72c37a005a96f1c77e8d35a8bc434"), privateKey.GeneratePublicKey())

	address, err := privateKey.GenerateAddress(nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, address.EncodeAddr(), "WEBD$gDWpzWQZgTPVu281pxPbiqWnT5Ti8t2p6b$")

	privateKey, err = NewPrivateKey(helpers.DecodeHex("b7598460fd0b21c0d4d542c4889c20346567041bcce48c4fddc80f6d2a026f94b69ac08d5e7ae0975d94e247c5247ab16a07bdf1092d3eb39bd1dd97d4b99188"))
	assert.NoError(t, err)

	assert.Equal(t, helpers.DecodeHex("b69ac08d5e7ae0975d94e247c5247ab16a07bdf1092d3eb39bd1dd97d4b99188"), privateKey.GeneratePublicKey())

	address, err = privateKey.GenerateAddress(nil, 0, nil)
	assert.NoError(t, err)

	assert.Equal(t, address.EncodeAddr(), "WEBD$gAJhXSnSv+$GgWVHFTPJZrnk@VyUaQqRkT$")
}
