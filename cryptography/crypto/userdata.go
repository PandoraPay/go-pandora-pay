package crypto

import (
	"errors"
	"golang.org/x/crypto/chacha20"
	"math/big"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
)

func EncryptDecryptUserData(key []byte, inputs ...[]byte) error {
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

// does an ECDH and generates a shared secret
// https://en.wikipedia.org/wiki/Elliptic_curve_Diffie%E2%80%93Hellman
func GenerateSharedSecret(secret *big.Int, peer_publickey *bn256.G1) ([]byte, error) {

	shared_point := new(bn256.G1).ScalarMult(peer_publickey, secret)
	compressed := shared_point.EncodeCompressed()
	if len(compressed) != 33 {
		return nil, errors.New("point compression needs to be fixed")
	}

	return cryptography.SHA3(compressed[:]), nil
}
