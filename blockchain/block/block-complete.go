package block

import (
	"bytes"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
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

func (blkComplete *BlockComplete) MerkleHash() helpers.Hash {
	var buffer = []byte{}
	if len(blkComplete.Txs) > 0 {

		//todo
		return cryptography.SHA3Hash(buffer)

	} else {
		return cryptography.SHA3Hash(buffer)
	}
}

func (blkComplete *BlockComplete) VerifyMerkleHash() bool {
	merkleHash := blkComplete.MerkleHash()
	return bytes.Equal(merkleHash[:], blkComplete.Block.MerkleHash[:])
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

func (blkComplete *BlockComplete) RemoveBlockComplete(accs *accounts.Accounts, toks *tokens.Tokens) {

	allFees := make(map[string]uint64)
	for i := len(blkComplete.Txs) - 1; i >= 0; i-- {
		blkComplete.Txs[i].AddFees(allFees)
		blkComplete.Txs[i].RemoveTransaction(blkComplete.Block.Height, accs, toks)
	}
	blkComplete.Block.RemoveBlock(accs, toks, allFees)
}

func (blkComplete *BlockComplete) Serialize() []byte {
	writer := helpers.NewBufferWriter()

	writer.Write(blkComplete.Block.Serialize())
	writer.WriteUvarint(uint64(len(blkComplete.Txs)))

	return writer.Bytes()
}

func (blkComplete *BlockComplete) Deserialize(buf []byte) {

	reader := helpers.NewBufferReader(buf)

	if uint64(len(buf)) > config.BLOCK_MAX_SIZE {
		panic("COMPLETE BLOCK EXCEEDS MAX SIZE")
	}

	blkComplete.Block.Deserialize(reader)

	txsCount := reader.ReadUvarint()
	//todo
	for i := uint64(0); i < txsCount; i++ {

	}

}
