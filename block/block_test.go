package block

import (
	"bytes"
	"pandora-pay/addresses"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	"testing"
	"time"
)

func TestBlock_Serialize(t *testing.T) {

	var err error

	privateKey := addresses.GenerateNewPrivateKey()
	publicKey, _ := privateKey.GeneratePublicKey()

	blockHeader := BlockHeader{MajorVersion: 0, MinorVersion: 0, Timestamp: uint64(time.Now().Unix()), Height: 0}
	block := Block{BlockHeader: blockHeader, MerkleHash: crypto.SHA3Hash([]byte("TEST")), Forger: publicKey[:], Signature: helpers.EmptyBytes(65)}

	buf := block.Serialize()
	if len(buf) < 30 {
		t.Errorf("Invalid serialization")
	}

	block2 := Block{}
	buf, err = block2.Deserialize(buf)
	if err != nil {
		t.Errorf("Final buff should be empty")
	}
	if len(buf) != 0 {
		t.Errorf("Final buff should be empty")
	}
	if !bytes.Equal(block2.Serialize(), block.Serialize()) {
		t.Errorf("Serialization/Deserialization doesn't work")
	}

}

func TestBlock_SerializeForSigning(t *testing.T) {

	var err error

	privateKey := addresses.GenerateNewPrivateKey()
	publicKey, _ := privateKey.GeneratePublicKey()

	blockHeader := BlockHeader{MajorVersion: 0, MinorVersion: 0, Timestamp: uint64(time.Now().Unix()), Height: 0}
	block := Block{BlockHeader: blockHeader, MerkleHash: crypto.SHA3Hash([]byte("TEST")), Forger: publicKey[:], Signature: helpers.EmptyBytes(65)}

	hash := block.SerializeForSigning()
	signature, err := privateKey.Sign(&hash)
	if err != nil {
		t.Errorf("Signing raised an error")
	}
	if len(signature) != 65 || bytes.Equal(signature, helpers.EmptyBytes(65)) {
		t.Errorf("Invalid signature")
	}
	block.Signature = signature

	if block.VerifySignature() != true {
		t.Errorf("Signature Validation failed")
	}

	if block.Signature[7] == 0 {
		block.Signature[7] = 5
	} else {
		block.Signature[7] = 0
	}

	if block.VerifySignature() != false {
		t.Errorf("Changed Signature Validation failed")
	}

}
