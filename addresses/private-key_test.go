package addresses

import (
	"bytes"
	"pandora-pay/helpers"
	"testing"
)

func TestGenerateNewPrivateKey(t *testing.T) {

	privateKey := GenerateNewPrivateKey()
	if len(privateKey.Key) != 32 {
		t.Errorf("Invalid private key length")
	}

	if bytes.Equal(privateKey.Key, helpers.EmptyBytes(32)) {
		t.Errorf("Invalid private key is empty")
	}

}

func TestPrivateKey_GenerateTransparentAddress(t *testing.T) {

	privateKey := GenerateNewPrivateKey()

	address := privateKey.GenerateTransparentAddress(false, 0, helpers.EmptyBytes(0))
	if len(address.PublicKey) != 33 || bytes.Equal(address.PublicKey, helpers.EmptyBytes(33)) {
		t.Errorf("Public Key is invalid")
	}

}
