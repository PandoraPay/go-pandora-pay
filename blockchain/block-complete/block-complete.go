package block_complete

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	merkle_tree "pandora-pay/cryptography/merkle-tree"
	"pandora-pay/helpers"
)

type BlockComplete struct {
	Block *block.Block
	Txs   []*transaction.Transaction
	Bloom *BlockCompleteBloom
}

func (blkComplete *BlockComplete) Validate() (err error) {
	if err = blkComplete.Block.Validate(); err != nil {
		return
	}
	for _, tx := range blkComplete.Txs {
		if err = tx.Validate(); err != nil {
			return
		}
	}
	return
}

func (blkComplete *BlockComplete) Verify() (err error) {
	if err = blkComplete.VerifyBloomAll(); err != nil {
		return
	}
	if err = blkComplete.Block.Verify(); err != nil {
		return
	}
	for _, tx := range blkComplete.Txs {
		if err = tx.Verify(); err != nil {
			return
		}
	}
	return
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

func (blkComplete *BlockComplete) VerifyMerkleHashManually() bool {
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

func (blkComplete *BlockComplete) Deserialize(buf []byte) (err error) {

	reader := helpers.NewBufferReader(buf)

	if uint64(len(buf)) > config.BLOCK_MAX_SIZE {
		return errors.New("COMPLETE BLOCK EXCEEDS MAX SIZE")
	}

	if err = blkComplete.Block.Deserialize(reader); err != nil {
		return
	}

	var txsCount uint64
	if txsCount, err = reader.ReadUvarint(); err != nil {
		return
	}

	blkComplete.Txs = make([]*transaction.Transaction, txsCount)
	for i := uint64(0); i < txsCount; i++ {
		blkComplete.Txs[i] = &transaction.Transaction{}
		if err = blkComplete.Txs[i].Deserialize(reader, true); err != nil {
			return
		}
	}

	return
}

func (blkComplete *BlockComplete) BloomAll(bloomTransactions, bloomBlock, bloomBlockComplete bool) (err error) {

	if bloomTransactions {
		for _, tx := range blkComplete.Txs {
			if err = tx.BloomAll(); err != nil {
				return
			}
		}
	}

	if bloomBlock {
		if err = blkComplete.Block.BloomNow(); err != nil {
			return
		}
	}

	if bloomBlockComplete {
		if err = blkComplete.BloomNow(); err != nil {
			return
		}
	}

	return
}

func CreateEmptyBlockComplete() *BlockComplete {
	return &BlockComplete{
		Block: &block.Block{
			BlockHeader: block.BlockHeader{
				Version: 0,
				Height:  0,
			},
		},
		Txs: []*transaction.Transaction{},
	}
}
