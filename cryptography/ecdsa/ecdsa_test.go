package ecdsa

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/helpers"
	"testing"
)

func TestPrivateKeyPublicKeyCreation(t *testing.T) {

	privateKey, err := GenerateKey()
	assert.Nil(t, err, "Error generating key")

	key := FromECDSA(privateKey)
	assert.Equal(t, len(key), 32, "Generatated Key length is invalid")

	publicKey, err := ComputePublicKey(key)
	assert.Nil(t, err, "Error generating key")
	assert.Equal(t, len(publicKey), 33, "Generatated Public Key Key length is invalid")

}

func TestECDSASignVerify(t *testing.T) {

	privateKey, err := GenerateKey()
	assert.Nil(t, err, "Error generating key")

	key := FromECDSA(privateKey)
	assert.Equal(t, len(key), 32, "Generatated Key length is invalid")

	message := helpers.RandomBytes(32)

	signature, err := Sign(message, privateKey)
	assert.Nil(t, err, "Error signing")
	assert.Equal(t, len(signature), 65, "Signing raised an error")

	signature = signature[0:64]

	emptySignature := helpers.EmptyBytes(64)
	assert.NotEqual(t, signature, emptySignature, "Signing is empty...")

	publicKey, err := ComputePublicKey(key)
	assert.Nil(t, err, "Error generating publickey")

	assert.Equal(t, VerifySignature(publicKey, message, signature), true, "Signature was not validated")
	assert.Equal(t, VerifySignature(publicKey, message, emptySignature), false, "Empty Signature was validated")

	var signature2 = make([]byte, len(signature))
	copy(signature2, signature)
	if signature2[2] == 5 {
		signature2[2] = 2
	} else {
		signature2[2] = 5
	}

	assert.Equal(t, VerifySignature(publicKey, message, signature2), false, "Changed Signature was validated")
}
