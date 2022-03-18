package pending_stakes_list

import (
	"pandora-pay/blockchain/data_storage/pending_stakes_list/pending_stakes"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type PendingStakesList struct {
	*hash_map.HashMap
}

func (self *PendingStakesList) CreateNewPendingStakes(blockHeight uint64) (*pending_stakes.PendingStakes, error) {
	key := strconv.FormatUint(blockHeight, 10)

	pendingStakes := pending_stakes.NewPendingStakes([]byte(key), 0) //index will be set by update
	pendingStakes.Height = blockHeight

	if err := self.Create(key, pendingStakes); err != nil {
		return nil, err
	}
	return pendingStakes, nil
}

func (self *PendingStakesList) GetPendingStakes(blockHeight uint64) (*pending_stakes.PendingStakes, error) {
	key := strconv.FormatUint(blockHeight, 10)

	data, err := self.Get(key)
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*pending_stakes.PendingStakes), nil
}

func NewPendingStakesList(tx store_db_interface.StoreDBTransactionInterface) (self *PendingStakesList) {

	hashmap := hash_map.CreateNewHashMap(tx, "pendingStakes", 0, false)

	self = &PendingStakesList{
		HashMap: hashmap,
	}

	self.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return pending_stakes.NewPendingStakes(key, index), nil
	}

	return
}
