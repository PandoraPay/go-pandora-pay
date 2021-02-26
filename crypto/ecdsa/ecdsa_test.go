package ecdsa

import (
	"bytes"
	"crypto/ecdsa"
	"pandora-pay/helpers"
	"testing"
)

func TestPrivateKeyPublicKeyCreation(t *testing.T) {

	var err error
	var privateKey *ecdsa.PrivateKey
	var publicKey []byte

	if privateKey, err = GenerateKey(); err != nil {
		t.Errorf("Generate Key failed %s", err)
	}

	key := FromECDSA(privateKey)
	if len(key) != 32 {
		t.Errorf("Generatated Key length is invalid %d", len(key))
	}

	if publicKey, err = ComputePublicKey(key); err != nil {
		t.Errorf("Generate Pub Key failed %s", err)
	}
	if len(publicKey) != 33 {
		t.Errorf("Generatated Key length is invalid %d", len(publicKey))
	}

}

func TestECDSASignVerify(t *testing.T) {

	var err error
	privateKey, _ := GenerateKey()

	key := FromECDSA(privateKey)
	if len(key) != 32 {
		t.Errorf("Generatated Key length is invalid %d", len(key))
	}

	message := helpers.RandomBytes(32)
	var signature []byte

	if signature, err = Sign(message, privateKey); err != nil {
		t.Errorf("Signing raised an error %s", err)
	}

	signature = signature[0:64]

	if len(signature) != 64 {
		t.Errorf("Signature length is invalid %d", len(signature))
	}

	emptySignature := helpers.EmptyBytes(64)
	if bytes.Equal(emptySignature, signature) {
		t.Errorf("Signature is empty %d", len(key))
	}

	publicKey, _ := ComputePublicKey(key)
	if !VerifySignature(publicKey, message, signature) {
		t.Errorf("Signature was not validated")
	}

	if VerifySignature(publicKey, message, emptySignature) {
		t.Errorf("Empty Signature was validated")
	}

	signature2 := signature[:]
	if signature2[2] == 5 {
		signature2[2] = 2
	} else {
		signature2[2] = 5
	}
	if VerifySignature(publicKey, message, signature2) {
		t.Errorf("Signature2 was validated")
	}
}
