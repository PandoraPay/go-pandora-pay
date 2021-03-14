package block

import (
	"bytes"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	merkle_tree "pandora-pay/cryptography/merkle-tree"
	"pandora-pay/helpers"
)

type BlockComplete struct {
	Block *Block
	Txs   []*transaction.Transaction
}

func (blkComplete *BlockComplete) Validate() {
	blkComplete.Block.Validate()
	for _, tx := range blkComplete.Txs {
		tx.Validate()
	}
}

func (blkComplete *BlockComplete) Verify() {
	blkComplete.Block.Verify()
	if blkComplete.VerifyMerkleHash() != true {
		panic("Verify Merkle Hash failed")
	}
	for _, tx := range blkComplete.Txs {
		tx.Verify()
	}
}

func (blkComplete *BlockComplete) MerkleHash() []byte {
	if len(blkComplete.Txs) > 0 {

		var hashes = make([][]byte, len(blkComplete.Txs))
		for i, tx := range blkComplete.Txs {
			hashes[i] = tx.ComputeHash()
		}
		return merkle_tree.MerkleRoot(hashes)
	} else {
		return cryptography.SHA3Hash([]byte{})
	}
}

func (blkComplete *BlockComplete) VerifyMerkleHash() bool {
	merkleHash := blkComplete.MerkleHash()
	return bytes.Equal(merkleHash, blkComplete.Block.MerkleHash)
}

func (blkComplete *BlockComplete) IncludeBlockComplete(accs *accounts.Accounts, toks *tokens.Tokens) {

	allFees := make(map[string]uint64)
	for _, tx := range blkComplete.Txs {
		tx.AddFees(allFees)
	}

	blkComplete.Block.IncludeBlock(accs, toks, allFees)

	for _, tx := range blkComplete.Txs {
		tx.IncludeTransaction(blkComplete.Block.Height, accs, toks)
	}

}

func (blkComplete *BlockComplete) Serialize() []byte {
	writer := helpers.NewBufferWriter()

	writer.Write(blkComplete.Block.Serialize())
	writer.WriteUvarint(uint64(len(blkComplete.Txs)))

	for _, tx := range blkComplete.Txs {
		serialized := tx.Serialize()
		writer.Write(serialized)
	}

	return writer.Bytes()
}

func (blkComplete *BlockComplete) Deserialize(buf []byte) {

	reader := helpers.NewBufferReader(buf)

	if uint64(len(buf)) > config.BLOCK_MAX_SIZE {
		panic("COMPLETE BLOCK EXCEEDS MAX SIZE")
	}

	blkComplete.Block.Deserialize(reader)

	txsCount := reader.ReadUvarint()
	blkComplete.Txs = make([]*transaction.Transaction, txsCount)
	for i := uint64(0); i < txsCount; i++ {
		txLength := reader.ReadUvarint()
		reader.ReadBytes(int(txLength))
		blkComplete.Txs[i] = &transaction.Transaction{}
		blkComplete.Txs[i].Deserialize(reader)
	}

}
