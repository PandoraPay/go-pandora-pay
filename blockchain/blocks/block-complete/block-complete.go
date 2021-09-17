package block_complete

import (
	"errors"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/data/accounts"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	merkle_tree "pandora-pay/cryptography/merkle-tree"
	"pandora-pay/helpers"
)

type BlockComplete struct {
	*block.Block
	Txs              []*transaction.Transaction `json:"txs"`
	BloomBlkComplete *BlockCompleteBloom        `json:"bloomBlkComplete"`
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

	if err = blkComplete.BloomBlkComplete.verifyIfBloomed(); err != nil {
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
		return cryptography.SHA3([]byte{})
	}
}

func (blkComplete *BlockComplete) IncludeBlockComplete(regs *registrations.Registrations, plainAccs *plain_accounts.PlainAccounts, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) (err error) {

	allFees := uint64(0)
	for _, tx := range blkComplete.Txs {
		fees := tx.ComputeFees()
		if err = helpers.SafeUint64Add(&allFees, fees); err != nil {
			return
		}
	}

	if err = blkComplete.Block.IncludeBlock(regs, plainAccs, accsCollection, toks, allFees); err != nil {
		return
	}

	for _, tx := range blkComplete.Txs {
		if err = tx.IncludeTransaction(blkComplete.Block.Height, regs, plainAccs, accsCollection, toks); err != nil {
			return
		}
	}

	return
}

func (blkComplete *BlockComplete) AdvancedSerialization(w *helpers.BufferWriter) {

	w.Write(blkComplete.Block.Bloom.Serialized)

	w.WriteUvarint(uint64(len(blkComplete.Txs)))

	for _, tx := range blkComplete.Txs {
		w.Write(tx.Bloom.Serialized)
	}
}

func (blkComplete *BlockComplete) Serialize(w *helpers.BufferWriter) {
	w.Write(blkComplete.BloomBlkComplete.Serialized)
}

func (blkComplete *BlockComplete) SerializeToBytes() []byte {
	return blkComplete.BloomBlkComplete.Serialized
}

func (blkComplete *BlockComplete) SerializeManualToBytes() []byte {
	writer := helpers.NewBufferWriter()
	blkComplete.AdvancedSerialization(writer)
	return writer.Bytes()
}

func (blkComplete *BlockComplete) Deserialize(r *helpers.BufferReader) (err error) {

	if uint64(len(r.Buf)) > config.BLOCK_MAX_SIZE {
		return errors.New("COMPLETE BLOCK EXCEEDS MAX SIZE")
	}

	first := r.Position

	if err = blkComplete.Block.Deserialize(r); err != nil {
		return
	}

	var txsCount uint64
	if txsCount, err = r.ReadUvarint(); err != nil {
		return
	}

	blkComplete.Txs = make([]*transaction.Transaction, txsCount)
	for i := uint64(0); i < txsCount; i++ {
		blkComplete.Txs[i] = &transaction.Transaction{}
		if err = blkComplete.Txs[i].Deserialize(r); err != nil {
			return
		}
	}

	//we can bloom more efficiently if asked
	blkComplete.BloomCompleteBySerialized(r.Buf[first:r.Position])

	return
}

func CreateEmptyBlockComplete() *BlockComplete {
	return &BlockComplete{
		Block: &block.Block{
			BlockHeader: &block.BlockHeader{},
		},
		Txs: []*transaction.Transaction{},
	}
}
