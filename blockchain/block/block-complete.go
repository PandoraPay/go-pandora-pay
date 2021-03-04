package block

import (
	"bytes"
	"errors"
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

	//todo
	for i := uint64(0); i < txsCount; i++ {

	}

	return
}
