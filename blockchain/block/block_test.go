package block

import (
	"bytes"
	"pandora-pay/addresses"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	"testing"
	"time"
)

var (
	merkleHash     = crypto.SHA3Hash([]byte("MerkleHash"))
	prevHash       = crypto.SHA3Hash([]byte("PrevHash"))
	prevKernelHash = crypto.SHA3Hash([]byte("PrevKernelHash"))
)

func TestBlock_Serialize(t *testing.T) {

	var err error

	privateKey := addresses.GenerateNewPrivateKey()
	publicKey, _ := privateKey.GeneratePublicKey()

	blockHeader := BlockHeader{Version: 0, Height: 0}
	blk := Block{
		BlockHeader:    blockHeader,
		MerkleHash:     merkleHash,
		PrevHash:       prevHash,
		PrevKernelHash: prevKernelHash,
		Timestamp:      uint64(time.Now().Unix()),
	}
	copy(blk.Forger[:], publicKey)

	buf := blk.Serialize()
	if len(buf) < 30 {
		t.Errorf("Invalid serialization")
	}

	blk2 := Block{}
	buf, err = blk2.Deserialize(buf)
	if err != nil {
		t.Errorf("Final buff should be empty")
	}
	if len(buf) != 0 {
		t.Errorf("Final buff should be empty")
	}
	if !bytes.Equal(blk2.Serialize(), blk.Serialize()) {
		t.Errorf("Serialization/Deserialization doesn't work")
	}

}

func TestBlock_SerializeForSigning(t *testing.T) {

	var err error

	privateKey := addresses.GenerateNewPrivateKey()
	publicKey, _ := privateKey.GeneratePublicKey()

	blockHeader := BlockHeader{Version: 0, Height: 0}
	blk := Block{
		BlockHeader:    blockHeader,
		MerkleHash:     merkleHash,
		PrevHash:       prevHash,
		PrevKernelHash: prevKernelHash,
		Timestamp:      uint64(time.Now().Unix()),
	}
	copy(blk.Forger[:], publicKey)

	hash := blk.SerializeForSigning()
	signature, err := privateKey.Sign(&hash)
	if err != nil {
		t.Errorf("Signing raised an error")
	}
	if len(signature) != 65 || bytes.Equal(signature, helpers.EmptyBytes(65)) {
		t.Errorf("Invalid signature")
	}
	copy(blk.Signature[:], signature)

	if blk.VerifySignature() != true {
		t.Errorf("Signature Validation failed")
	}

	if blk.Signature[7] == 0 {
		blk.Signature[7] = 5
	} else {
		blk.Signature[7] = 0
	}

	if blk.VerifySignature() != false {
		t.Errorf("Changed Signature Validation failed")
	}

}