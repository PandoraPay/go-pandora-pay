package block

import (
	"bytes"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type BlockComplete struct {
	Block *Block
	Txs   []*transaction.Transaction
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
