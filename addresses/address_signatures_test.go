package addresses

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
)

func Test_VerifySignedMessage(t *testing.T) {

	for i := 0; i < 100; i++ {

		privateKey := GenerateNewPrivateKey()
		address, err := privateKey.GenerateAddress(false, 0, nil)
		assert.Nil(t, err, "Error generating key")

		message := helpers.RandomBytes(32)
		signature, err := privateKey.Sign(message)
		assert.Nil(t, err, "Error signing")

		assert.Equal(t, len(signature), cryptography.SignatureSize, "signature length is invalid")

		emptySignature := helpers.EmptyBytes(cryptography.SignatureSize)
		assert.NotEqual(t, signature, emptySignature, "Signing is empty...")

		assert.Equal(t, address.VerifySignedMessage(message, signature), true, "verification failed")

		var signature2 = helpers.CloneBytes(signature)
		copy(signature2, signature)
		if signature2[2] == 5 {
			signature2[2] = 2
		} else {
			signature2[2] = 5
		}

		assert.Equal(t, address.VerifySignedMessage(message, signature2), false, "Changed Signature was validated")

	}

}
