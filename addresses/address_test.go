package addresses

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/custom_base64"
	"testing"
)

func TestAddress_EncodeAddr(t *testing.T) {

	//WIF
	//1+33+1+4

	privateKey := GenerateNewPrivateKey()
	address, err := privateKey.GenerateAddress(false, nil, false, helpers.EmptyBytes(0), 0, helpers.EmptyBytes(0))
	assert.NoError(t, err)
	assert.Equal(t, len(address.PublicKey), cryptography.PublicKeySize)
	assert.Equal(t, len(address.PaymentID), 0)
	assert.Equal(t, len(address.PaymentAsset), 0)
	assert.Equal(t, len(address.SpendPublicKey), 0)
	assert.Equal(t, address.Staked, false)
	assert.Equal(t, len(address.Registration), 0)

	encoded := address.EncodeAddr()

	decoded, err := custom_base64.Base64Encoder.DecodeString(encoded[config.NETWORK_BYTE_PREFIX_LENGTH:])
	assert.NoError(t, err, "Address Decoding raised an error")
	assert.Equal(t, len(decoded), 1+cryptography.PublicKeySize+1+4, "AddressEncoded length is invalid")

	address, err = privateKey.GenerateAddress(false, nil, false, helpers.EmptyBytes(0), 20, helpers.EmptyBytes(0))
	assert.NoError(t, err)
	assert.Equal(t, len(address.PublicKey), cryptography.PublicKeySize)
	assert.Equal(t, len(address.PaymentID), 0)
	assert.Equal(t, len(address.PaymentAsset), 0)
	assert.Equal(t, len(address.SpendPublicKey), 0)
	assert.Equal(t, address.PaymentAmount, uint64(20))
	assert.Equal(t, address.Staked, false)
	assert.Equal(t, len(address.Registration), 0)

	encodedAmount := address.EncodeAddr()
	assert.NotEqual(t, len(encoded), len(encodedAmount), "Encoded Amounts are invalid")
	assert.NotEqual(t, encoded, encodedAmount, "Encoded Amounts are invalid")

	address, err = privateKey.GenerateAddress(false, nil, false, helpers.EmptyBytes(8), 20, nil)
	assert.NoError(t, err)
	assert.Equal(t, len(address.PublicKey), cryptography.PublicKeySize)
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

		var spendPublicKey, paymentID, paymentAsset []byte
		staked := rand.Intn(2) == 0
		registration := rand.Intn(2) == 0
		if rand.Intn(2) == 0 {
			spendPublicKey = helpers.RandomBytes(cryptography.PublicKeySize)
		}
		if rand.Intn(2) == 0 {
			paymentID = helpers.RandomBytes(8)
		}
		if rand.Intn(2) == 0 {
			paymentAsset = helpers.EmptyBytes(config_coins.ASSET_LENGTH)
		}
		paymentAmount := rand.Uint64()

		privateKey := GenerateNewPrivateKey()
		address, err := privateKey.GenerateAddress(staked, spendPublicKey, registration, paymentID, paymentAmount, paymentAsset)
		assert.NoError(t, err)

		encoded := address.EncodeAddr()

		decodedAddress, err := DecodeAddr(encoded)
		assert.NoError(t, err, "Invalid Decoded Address")

		assert.Equal(t, decodedAddress.PublicKey, address.PublicKey)
		assert.Equal(t, decodedAddress.PaymentAmount, paymentAmount)
		assert.Equal(t, bytes.Equal(decodedAddress.PaymentID, paymentID), true)
		assert.Equal(t, bytes.Equal(decodedAddress.PaymentAsset, paymentAsset), true)
		assert.Equal(t, bytes.Equal(decodedAddress.SpendPublicKey, spendPublicKey), true)
		if registration {
			assert.Equal(t, len(decodedAddress.Registration), cryptography.SignatureSize)
		} else {
			assert.Equal(t, len(decodedAddress.Registration), 0)
		}
		assert.Equal(t, decodedAddress.Staked, staked)

		encoded2 := decodedAddress.EncodeAddr()

		assert.Equal(t, encoded2, encoded, "Encoded addresses are not matching")

	}

}
