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

func TestPrivateKey_GenerateAddress(t *testing.T) {

	privateKey := GenerateNewPrivateKey()

	address, err := privateKey.GenerateAddress(false, 0, helpers.EmptyBytes(0))
	if err != nil {
		t.Errorf("Address Generation raised an error")
	}

	if len(address.PublicKey) != 33 || bytes.Equal(address.PublicKey, helpers.EmptyBytes(33)) ||
		address.Amount != 0 || len(address.PaymentID) != 0 {
		t.Errorf("Generated Address is invalid")
	}

	address, err = privateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	if len(address.PublicKey) != 20 || err != nil || address.Amount != 0 || len(address.PaymentID) != 0 {
		t.Errorf("Generated Address is invalid")
	}

	address, err = privateKey.GenerateAddress(true, 20, helpers.RandomBytes(8))
	if len(address.PublicKey) != 20 || err != nil || address.Amount != 20 || len(address.PaymentID) != 8 {
		t.Errorf("Generated Address is invalid")
	}

}
