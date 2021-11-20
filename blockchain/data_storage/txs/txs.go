package txs

import (
	"pandora-pay/cryptography"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Txs struct {
	*hash_map.HashMap `json:"-"`
}

func (txs *Txs) CreateNewTx(hash, tx []byte) (*TxStorage, error) {
	blockStorage := NewTxStorage(hash, 0, tx) //index will be set by update
	if err := txs.Create(string(hash), blockStorage); err != nil {
		return nil, err
	}
	return blockStorage, nil
}

func (txs *Txs) GetBlock(key []byte) (*TxStorage, error) {

	data, err := txs.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*TxStorage), nil
}

func NewTxs(tx store_db_interface.StoreDBTransactionInterface) (txs *Txs) {

	hashmap := hash_map.CreateNewHashMap(tx, "txs", cryptography.HashSize, true)

	txs = &Txs{
		HashMap: hashmap,
	}

	txs.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return NewTxStorage(key, index, nil), nil
	}

	return
}
