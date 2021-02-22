package merkle_tree

import (
	"bytes"
	"pandora-pay/crypto"
	"testing"
)

func TestMerkleRoot(t *testing.T) {

	var hashes [2]Hash
	hashes[0] = crypto.RandomHash()
	hashes[1] = crypto.RandomHash()

	root := MerkleRoot(hashes[:])

	hash := *hashMerkleNode(&hashes[0], &hashes[1])
	if !bytes.Equal(root[:], hash[:]) {
		t.Errorf("Merkle Tree Hashes are invalid %s %s %s", string(hashes[0][:]), string(hashes[1][:]), string(root[:]))
	}

}
