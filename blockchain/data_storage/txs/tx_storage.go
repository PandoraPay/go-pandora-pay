package txs

import (
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type TxStorage struct {
	hash_map.HashMapElementSerializableInterface
	Height uint64
	Hash   []byte
	Tx     []byte
}

func (txStorage *TxStorage) SetIndex(index uint64) {
	txStorage.Height = index
}

func (txStorage *TxStorage) SetKey(key []byte) {
	txStorage.Hash = key
}

func (txStorage *TxStorage) GetIndex() uint64 {
	return txStorage.Height
}

func (txStorage *TxStorage) Validate() error {
	return nil
}

func (txStorage *TxStorage) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(uint64(len(txStorage.Tx)))
	w.Write(txStorage.Tx)
}

func (txStorage *TxStorage) Deserialize(r *helpers.BufferReader) (err error) {
	var n uint64

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if txStorage.Tx, err = r.ReadBytes(int(n)); err != nil {
		return
	}

	return
}

func NewTxStorage(hash []byte, height uint64, tx []byte) *TxStorage {
	return &TxStorage{
		nil,
		height,
		hash,
		tx,
	}
}
