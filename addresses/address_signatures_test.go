package addresses

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
)

func Test_VerifySignedMessage(t *testing.T) {

	for i := 0; i < 100; i++ {

		privateKey := GenerateNewPrivateKey()

		message := helpers.RandomBytes(32)
		signature, err := privateKey.Sign(message)
		assert.Nil(t, err, "Error signing")

		assert.Equal(t, len(signature), cryptography.SignatureSize, "signature length is invalid")

		emptySignature := helpers.EmptyBytes(cryptography.SignatureSize)
		assert.Equal(t, bytes.Equal(signature, emptySignature), false, "Signing is empty...")

		assert.Equal(t, privateKey.Verify(message, signature), true, "verification failed")

		var signature2 = helpers.CloneBytes(signature)
		copy(signature2, signature)

		value := byte(rand.Uint64() % 256)
		if signature2[2] == value {
			signature2[2] = value + byte(rand.Uint64()%255)
		} else {
			signature2[2] = value
		}

		assert.Equal(t, privateKey.Verify(message, signature2), false, "Changed Signature was validated")

	}

}
