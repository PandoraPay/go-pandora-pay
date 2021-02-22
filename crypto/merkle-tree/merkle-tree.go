package merkle_tree

import (
	"math"
	"pandora-pay/crypto"
)

/**
Fast Merkle Tree Construction
It can't generate efficient proofs
*/

type Hash = crypto.Hash

func roundNextPowerOfTwo(number int) int {

	if number&(number-1) == 0 {
		return number
	}

	exp := uint(math.Log2(float64(number))) + 1
	return 1 << exp
}

func hashMerkleNode(left *Hash, right *Hash) *Hash {
	// Concatenate the left and right nodes.
	var hash [len(left) * 2]byte
	copy(hash[:len(left)], left[:])
	copy(hash[len(left):], right[:])

	hash2 := crypto.SHA3Hash(hash[:])
	return &hash2
}

func buildMerkleTree(hashes []Hash) []*Hash {

	if len(hashes) == 0 {
		merkles := make([]*Hash, 1)
		merkles[0] = &Hash{}
		return merkles
	}

	roundedNextPowerOfTwo := roundNextPowerOfTwo(len(hashes))
	arraySize := roundedNextPowerOfTwo*2 - 1
	nodes := make([]*Hash, arraySize)

	for i := range hashes {
		nodes[i] = &hashes[i]
	}

	offset := roundedNextPowerOfTwo
	for i := 0; i < arraySize-1; i += 2 {

		switch {

		case nodes[i] == nil:
			nodes[offset] = nil

		case nodes[i+1] == nil:
			newHash := hashMerkleNode(nodes[i], nodes[i])
			nodes[offset] = newHash

		default:
			newHash := hashMerkleNode(nodes[i], nodes[i+1])
			nodes[offset] = newHash
		}
		offset++
	}
	return nodes
}

func MerkleRoot(hashes []Hash) Hash {
	merkles := buildMerkleTree(hashes)
	return *merkles[len(merkles)-1] //return last element
}
