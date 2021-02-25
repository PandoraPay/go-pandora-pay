package block

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	"pandora-pay/transaction"
)

type BlockComplete struct {
	Block *Block
	Txs   []*transaction.Transaction
}

func (blkComplete *BlockComplete) MerkleHash() crypto.Hash {

	var buffer = []byte{}

	if len(blkComplete.Txs) > 0 {

		//todo
		return crypto.SHA3Hash(buffer)

	} else {
		return crypto.SHA3Hash(buffer)
	}

}

func (blkComplete *BlockComplete) VerifyMerkleHash() bool {

	merkleHash := blkComplete.MerkleHash()
	return bytes.Equal(merkleHash[:], blkComplete.Block.MerkleHash[:])

}

func (blkComplete *BlockComplete) Serialize() []byte {

	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	blockSerialized := blkComplete.Block.Serialize()
	serialized.Write(blockSerialized)

	n := binary.PutUvarint(buf, uint64(len(blkComplete.Txs)))
	serialized.Write(buf[:n])

	return serialized.Bytes()
}

func (blkComplete *BlockComplete) Deserialize(buf []byte) (out []byte, err error) {

	if uint64(len(buf)) > config.BLOCK_MAX_SIZE {
		err = errors.New("COMPLETE BLOCK EXCEEDS MAX SIZE")
		return
	}

	out = buf

	out, err = blkComplete.Block.Deserialize(out)
	if err != nil {
		return
	}

	var txsCount uint64
	txsCount, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}

	//todo
	for i := uint64(0); i < txsCount; i++ {

	}

	buf = out
	return
}
