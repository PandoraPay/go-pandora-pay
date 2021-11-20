package blocks

import (
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type BlockStorage struct {
	hash_map.HashMapElementSerializableInterface
	Height uint64
	Hash   []byte
	Block  []byte
	Txs    [][]byte
}

func (blockStorage *BlockStorage) SetIndex(index uint64) {
	blockStorage.Height = index
}

func (blockStorage *BlockStorage) SetKey(key []byte) {
	blockStorage.Hash = key
}

func (blockStorage *BlockStorage) GetIndex() uint64 {
	return blockStorage.Height
}

func (blockStorage *BlockStorage) Validate() error {
	return nil
}

func (blockStorage *BlockStorage) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(uint64(len(blockStorage.Block)))
	w.Write(blockStorage.Block)
	w.WriteUvarint(uint64(len(blockStorage.Txs)))
	for _, tx := range blockStorage.Txs {
		w.Write(tx)
	}
}

func (blockStorage *BlockStorage) Deserialize(r *helpers.BufferReader) (err error) {
	var n uint64

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if blockStorage.Block, err = r.ReadBytes(int(n)); err != nil {
		return
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	blockStorage.Txs = make([][]byte, n)
	for i := uint64(0); i < n; i++ {
		if blockStorage.Txs[i], err = r.ReadHash(); err != nil {
			return
		}
	}

	return
}

func NewBlockStorage(hash []byte, height uint64, block []byte, txs [][]byte) *BlockStorage {
	return &BlockStorage{
		nil,
		height,
		hash,
		block,
		txs,
	}
}
