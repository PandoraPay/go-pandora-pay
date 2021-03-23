package block

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"testing"
	"time"
)

var (
	merkleHash     = cryptography.SHA3Hash([]byte("MerkleHash"))
	prevHash       = cryptography.SHA3Hash([]byte("PrevHash"))
	prevKernelHash = cryptography.SHA3Hash([]byte("PrevKernelHash"))
)

func TestBlock_Serialize(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	publicKey, err := privateKey.GeneratePublicKey()
	assert.NoError(t, err, "Error...?")

	publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)

	blockHeader := BlockHeader{Version: 0, Height: 0}
	blk := Block{
		BlockHeader:    blockHeader,
		MerkleHash:     merkleHash,
		PrevHash:       prevHash,
		PrevKernelHash: prevKernelHash,
		Forger:         publicKeyHash,
		Timestamp:      uint64(time.Now().Unix()),
		Signature:      make([]byte, 65),
	}

	buf := blk.Serialize()
	assert.Equal(t, len(buf) < 30, false, "Invalid serialization")

	blk2 := Block{}

	reader := helpers.NewBufferReader(buf)
	err = blk2.Deserialize(reader)
	assert.NoError(t, err, "Error...?")

	assert.Equal(t, blk2.Serialize(), blk.Serialize(), "Serialization/Deserialization doesn't work")

}

func TestBlock_SerializeForSigning(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	publicKey, err := privateKey.GeneratePublicKey()
	assert.NoError(t, err, "Error...?")

	publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)

	blockHeader := BlockHeader{Version: 0, Height: 0}
	blk := Block{
		BlockHeader:    blockHeader,
		MerkleHash:     merkleHash,
		PrevHash:       prevHash,
		PrevKernelHash: prevKernelHash,
		Forger:         publicKeyHash,
		Timestamp:      uint64(time.Now().Unix()),
		Signature:      make([]byte, 65),
	}

	hash := blk.SerializeForSigning()
	var signature []byte

	signature, err = privateKey.Sign(hash)
	assert.NoError(t, err, "Signing raised an error")

	assert.NotEqual(t, signature, helpers.EmptyBytes(65), "Invalid signature")
	blk.Signature = signature

	assert.Equal(t, blk.VerifySignatureManually(), true, "Signature Validation failed")

	if blk.Signature[7] == 0 {
		blk.Signature[7] = 5
	} else {
		blk.Signature[7] = 0
	}

	assert.Equal(t, blk.VerifySignatureManually(), false, "Changed Signature Validation failed")

}
