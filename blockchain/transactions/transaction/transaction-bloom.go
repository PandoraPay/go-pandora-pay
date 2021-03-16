package transaction

import "pandora-pay/cryptography"

type TransactionBloom struct {
	Serialized []byte
	Size       uint64
	Hash       []byte
	HashStr    string
	bloomed    bool
}

func (tx *Transaction) BloomNow() {
	tx.Bloom = new(TransactionBloom)
	tx.Bloom.BloomTransactionHash(tx)
}

func (bloom *TransactionBloom) BloomTransactionHash(tx *Transaction) {
	bloom.Serialized = tx.Serialize()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.Hash = cryptography.SHA3Hash(bloom.Serialized)
	bloom.HashStr = string(bloom.Hash)
	bloom.bloomed = true
}

func (bloom *TransactionBloom) VerifyIfBloomed() {
	if !bloom.bloomed {
		panic("Tx is not bloomed")
	}
}
