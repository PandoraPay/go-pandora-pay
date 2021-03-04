package merkle_tree

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/crypto"
	"testing"
)

func TestMerkleRoot(t *testing.T) {

	var hashes [2]Hash
	hashes[0] = crypto.RandomHash()
	hashes[1] = crypto.RandomHash()

	root := MerkleRoot(hashes[:])

	hash := *hashMerkleNode(&hashes[0], &hashes[1])
	assert.Equal(t, root, hash, "Merkle Tree Hashes are invalid")

}
