package addresses

import (
	"bytes"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
)

func TestAddress_EncodeAddr(t *testing.T) {

	//WIF
	//1+33+1+4

	privateKey := GenerateNewPrivateKey()
	address, err := privateKey.GenerateAddress(helpers.EmptyBytes(0), 0, helpers.EmptyBytes(0))
	assert.NoError(t, err)
	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.Equal(t, len(address.PaymentID), 0)
	assert.Equal(t, len(address.PaymentAsset), 0)

	encoded := address.EncodeAddr()

	decoded, err := base58.Decode(encoded[config.NETWORK_BYTE_PREFIX_LENGTH:])
	assert.NoError(t, err, "Address Decoding raised an error")
	assert.Equal(t, len(decoded), 1+cryptography.PublicKeyHashSize+1+4, "AddressEncoded length is invalid")

	address, err = privateKey.GenerateAddress(helpers.EmptyBytes(0), 1235312323551220, helpers.EmptyBytes(0))
	assert.NoError(t, err)
	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.Equal(t, len(address.PaymentID), 0)
	assert.Equal(t, len(address.PaymentAsset), 0)
	assert.Equal(t, address.PaymentAmount, uint64(1235312323551220))

	encodedAmount := address.EncodeAddr()
	assert.NotEqual(t, encoded, encodedAmount, "Encoded Amounts are invalid")

	address, err = privateKey.GenerateAddress(helpers.EmptyBytes(8), 20, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(address.PublicKeyHash), cryptography.PublicKeyHashSize)
	assert.Equal(t, len(address.PaymentID), 8)
	assert.Equal(t, address.PaymentAmount, uint64(20))

	encodedAmountPaymentID := address.EncodeAddr()
	assert.Nil(t, err, "Encoding Address raised an error")
	assert.NotEqual(t, len(encodedAmount), len(encodedAmountPaymentID))
	assert.NotEqual(t, len(encoded), len(encodedAmountPaymentID))
	assert.NotEqual(t, encodedAmount, encodedAmountPaymentID)
	assert.NotEqual(t, encoded, encodedAmountPaymentID)

}

func TestDecodeAddr(t *testing.T) {

	for i := 0; i < 100; i++ {

		var paymentID, paymentAsset []byte
		if rand.Intn(2) == 0 {
			paymentID = helpers.RandomBytes(8)
		}
		if rand.Intn(2) == 0 {
			paymentAsset = helpers.EmptyBytes(config_coins.ASSET_LENGTH)
		}
		paymentAmount := rand.Uint64()

		privateKey := GenerateNewPrivateKey()
		address, err := privateKey.GenerateAddress(paymentID, paymentAmount, paymentAsset)
		assert.NoError(t, err)

		encoded := address.EncodeAddr()

		decodedAddress, err := DecodeAddr(encoded)
		assert.NoError(t, err, "Invalid Decoded Address")

		assert.Equal(t, decodedAddress.PublicKeyHash, address.PublicKeyHash)
		assert.Equal(t, decodedAddress.PaymentAmount, paymentAmount)
		assert.Equal(t, bytes.Equal(decodedAddress.PaymentID, paymentID), true)
		assert.Equal(t, bytes.Equal(decodedAddress.PaymentAsset, paymentAsset), true)

		encoded2 := decodedAddress.EncodeAddr()

		assert.Equal(t, encoded2, encoded, "Encoded addresses are not matching")

	}

}
