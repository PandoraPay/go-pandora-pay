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

	blockHeader := BlockHeader{Version: 0, Timestamp: uint64(time.Now().Unix()), Height: 0}
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
