package crypto

import (
	"bytes"
	"pandora-pay/crypto/ecdsa"
	"pandora-pay/helpers"
	"testing"
)

func PrivateKeyPublicKeyCreation(t *testing.T) {

	privateKey, err := ecdsa.GenerateKey()
	if err != nil {
		t.Errorf("Generate Key failed %s", err)
	}

	key := ecdsa.FromECDSA(privateKey)
	if len(key) != 32 {
		t.Errorf("Generatated Key length is invalid %d", len(key))
	}

	publicKey, err := ComputePublicKey(key)
	if err != nil {
		t.Errorf("Generate Pub Key failed %s", err)
	}
	if len(publicKey) != 33 {
		t.Errorf("Generatated Key length is invalid %d", len(publicKey))
	}

}

func ECDSASignVerify(t *testing.T) {

	privateKey, _ := ecdsa.GenerateKey()

	key := ecdsa.FromECDSA(privateKey)
	if len(key) != 32 {
		t.Errorf("Generatated Key length is invalid %d", len(key))
	}

	message := helpers.RandomBytes(40)

	signature, err := ecdsa.Sign(SHA3(message), privateKey)
	if err != nil {
		t.Errorf("Signing raised an error %s", err)
	}

	if len(signature) != 65 {
		t.Errorf("Signature length is invalid %d", len(key))
	}

	emptySignature := helpers.EmptyBytes(65)
	if bytes.Equal(emptySignature, signature) {
		t.Errorf("Signature is empty %d", len(key))
	}

	publicKey, _ := ComputePublicKey(key)
	if !ecdsa.VerifySignature(publicKey, SHA3(message), signature) {
		t.Errorf("Signature was not validated")
	}

	if ecdsa.VerifySignature(publicKey, SHA3(message), emptySignature) {
		t.Errorf("Empty Signature was validated")
	}

	signature2 := signature[:]
	if signature2[2] == 5 {
		signature2[2] = 2
	} else {
		signature2[2] = 5
	}
	if ecdsa.VerifySignature(publicKey, SHA3(message), signature2) {
		t.Errorf("Signature2 was validated")
	}
}
