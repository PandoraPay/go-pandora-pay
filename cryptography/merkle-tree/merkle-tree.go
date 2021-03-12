package merkle_tree

import (
	"math"
	"pandora-pay/cryptography"
)

/**
Fast Merkle Tree Construction
It can't generate efficient proofs
*/

func roundNextPowerOfTwo(number int) int {

	if number&(number-1) == 0 {
		return number
	}

	exp := uint(math.Log2(float64(number))) + 1
	return 1 << exp
}

func hashMerkleNode(left []byte, right []byte) []byte {
	// Concatenate the left and right nodes.
	hash := append(left, right...)
	return cryptography.SHA3Hash(hash)
}

func buildMerkleTree(hashes [][]byte) [][]byte {

	if len(hashes) == 0 {
		return [][]byte{}
	}

	roundedNextPowerOfTwo := roundNextPowerOfTwo(len(hashes))
	arraySize := roundedNextPowerOfTwo*2 - 1
	nodes := make([][]byte, arraySize)

	for i := range hashes {
		nodes[i] = hashes[i]
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

func MerkleRoot(hashes [][]byte) []byte {
	merkles := buildMerkleTree(hashes)
	return merkles[len(merkles)-1] //return last element
}
