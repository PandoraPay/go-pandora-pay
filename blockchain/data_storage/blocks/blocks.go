package blocks

import (
	"pandora-pay/cryptography"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Blocks struct {
	*hash_map.HashMap `json:"-"`
}

func (blocks *Blocks) CreateNewBlock(hash []byte, block []byte, txs [][]byte) (*BlockStorage, error) {
	blockStorage := NewBlockStorage(hash, 0, block, txs) //index will be set by update
	if err := blocks.Create(string(hash), blockStorage); err != nil {
		return nil, err
	}
	return blockStorage, nil
}

func (blocks *Blocks) GetBlock(key []byte) (*BlockStorage, error) {

	data, err := blocks.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*BlockStorage), nil
}

func NewBlocks(tx store_db_interface.StoreDBTransactionInterface) (blocks *Blocks) {

	hashmap := hash_map.CreateNewHashMap(tx, "blocks", cryptography.HashSize, true)

	blocks = &Blocks{
		HashMap: hashmap,
	}

	blocks.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return NewBlockStorage(key, index, nil, nil), nil
	}

	return
}
