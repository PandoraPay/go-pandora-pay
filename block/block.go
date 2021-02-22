package block

import (
	"pandora-pay/crypto"
)

type Block struct {
	BlockHeader
	MerkleHash crypto.Hash

	Forger    [33]byte // public key
	Signature [65]byte // signature

}

type BlockComplete struct {
	block *Block
	//txs []*transaction
}

func (block *Block) ComputeHash() (crypto.Hash, error) {

}

func (block *Block) ComputeKernelHash() (crypto.Hash, error) {

}

func (block *Block) Serialize() ([]byte, error) {

}

func (block *Block) Deserialize(b []byte) error {

}
