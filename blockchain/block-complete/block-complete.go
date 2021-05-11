package block_complete

import (
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
	helpers.SerializableInterface
	Block *block.Block
	Txs   []*transaction.Transaction
	Bloom *BlockCompleteBloom `json:"-"`
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
			hashes[i] = tx.Bloom.Hash
		}
		return merkle_tree.MerkleRoot(hashes)
	} else {
		return cryptography.SHA3Hash([]byte{})
	}
}

func (blkComplete *BlockComplete) IncludeBlockComplete(accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	allFees := make(map[string]uint64)
	for _, tx := range blkComplete.Txs {
		if err = tx.AddFees(allFees); err != nil {
			return
		}
	}

	if err = blkComplete.Block.IncludeBlock(accs, toks, allFees); err != nil {
		return
	}

	for _, tx := range blkComplete.Txs {
		if err = tx.IncludeTransaction(blkComplete.Block.Height, accs, toks); err != nil {
			return
		}
	}

	return
}

func (blkComplete *BlockComplete) Serialize(writer *helpers.BufferWriter) {

	writer.Write(blkComplete.Block.SerializeToBytes())
	writer.WriteUvarint(uint64(len(blkComplete.Txs)))

	for _, tx := range blkComplete.Txs {
		writer.Write(tx.Bloom.Serialized)
	}

}

func (blkComplete *BlockComplete) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	blkComplete.Serialize(writer)
	return writer.Bytes()
}

func (blkComplete *BlockComplete) Deserialize(reader *helpers.BufferReader) (err error) {

	if uint64(len(reader.Buf)) > config.BLOCK_MAX_SIZE {
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
		if err = blkComplete.Txs[i].Deserialize(reader); err != nil {
			return
		}
	}

	return
}

func (blkComplete *BlockComplete) BloomAll() (err error) {

	for _, tx := range blkComplete.Txs {
		if err = tx.BloomAll(); err != nil {
			return
		}
	}

	if err = blkComplete.Block.BloomNow(); err != nil {
		return
	}
	if err = blkComplete.BloomNow(); err != nil {
		return
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
