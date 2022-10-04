package conditional_payments_list

import (
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type ConditionalPaymentsCollection struct {
	tx   store_db_interface.StoreDBTransactionInterface
	maps map[string]*ConditionalPaymentsHashMap
	list []hash_map.HashMapInterface
}

func (collection *ConditionalPaymentsCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
}

func (this *ConditionalPaymentsCollection) GetAllMaps() map[string]*ConditionalPaymentsHashMap {
	return this.maps
}

func (this *ConditionalPaymentsCollection) GetAllHashmaps() []hash_map.HashMapInterface {
	return this.list
}

func (this *ConditionalPaymentsCollection) GetMap(blockHeight uint64) (*ConditionalPaymentsHashMap, error) {

	it := this.maps[string(blockHeight)]
	if it == nil {
		it = NewConditionalPaymentsHashMap(this.tx, blockHeight)
		this.list = append(this.list, it.HashMap)
		this.maps[string(blockHeight)] = it
	}

	return it, nil
}

func NewConditionalPaymentsCollection(tx store_db_interface.StoreDBTransactionInterface) *ConditionalPaymentsCollection {
	return &ConditionalPaymentsCollection{
		tx,
		make(map[string]*ConditionalPaymentsHashMap),
		make([]hash_map.HashMapInterface, 0),
	}
}
