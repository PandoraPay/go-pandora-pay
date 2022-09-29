package pending_future_list

import (
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type PendingFutureCollection struct {
	tx   store_db_interface.StoreDBTransactionInterface
	maps map[string]*PendingFutureHashMap
	list []hash_map.HashMapInterface
}

func (collection *PendingFutureCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
}

func (this *PendingFutureCollection) GetAllMaps() map[string]*PendingFutureHashMap {
	return this.maps
}

func (this *PendingFutureCollection) GetAllHashmaps() []hash_map.HashMapInterface {
	return this.list
}

func (this *PendingFutureCollection) GetMap(blockHeight uint64) (*PendingFutureHashMap, error) {

	it := this.maps[string(blockHeight)]
	if it == nil {
		it = NewPendingFutureHashMap(this.tx, blockHeight)
		this.list = append(this.list, it.HashMap)
		this.maps[string(blockHeight)] = it
	}

	return it, nil
}

func NewPendingFutureCollection(tx store_db_interface.StoreDBTransactionInterface) *PendingFutureCollection {
	return &PendingFutureCollection{
		tx,
		make(map[string]*PendingFutureHashMap),
		make([]hash_map.HashMapInterface, 0),
	}
}
