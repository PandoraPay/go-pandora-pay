package merkle_tree

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/cryptography"
	"testing"
)

func TestMerkleRoot(t *testing.T) {

	var hashes [][]byte
	hashes = append(hashes, cryptography.RandomHash())
	hashes = append(hashes, cryptography.RandomHash())

	root := MerkleRoot(hashes)

	hash := hashMerkleNode(hashes[0], hashes[1])
	assert.Equal(t, root, hash, "Merkle Tree Hashes are invalid")

}
