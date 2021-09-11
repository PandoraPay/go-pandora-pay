package crypto

import (
	"errors"
	"golang.org/x/crypto/chacha20"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
)

func EncryptDecryptUserData(blinder *bn256.G1, inputs ...[]byte) error {
	blinder_compressed := blinder.EncodeCompressed()
	if len(blinder_compressed) != 33 {
		return errors.New("point compression needs to be fixed")
	}

	key := cryptography.SHA3(blinder_compressed[:])
	var nonce [24]byte // nonce is 24 bytes, we will use xchacha20

	cipher, err := chacha20.NewUnauthenticatedCipher(key[:], nonce[:])
	if err != nil {
		return err
	}

	for _, input := range inputs {
		cipher.XORKeyStream(input, input)
	}
	return nil
}
