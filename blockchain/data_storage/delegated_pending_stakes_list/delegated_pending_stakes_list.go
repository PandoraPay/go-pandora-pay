package delegated_pending_stakes_list

import (
	"pandora-pay/blockchain/data_storage/delegated_pending_stakes_list/delegated_pending_stakes"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type DelegatedPendingStakesList struct {
	*hash_map.HashMap
}

func (delegatedPendingStakesList *DelegatedPendingStakesList) CreateNewDelegatedPendingStakes(blockHeight uint64) (*delegated_pending_stakes.DelegatedPendingStakes, error) {
	key := strconv.FormatUint(blockHeight, 10)

	delegatedPendingStakes := delegated_pending_stakes.NewDelegatedPendingStakes([]byte(key), 0) //index will be set by update
	delegatedPendingStakes.Height = blockHeight

	if err := delegatedPendingStakesList.Create(key, delegatedPendingStakes); err != nil {
		return nil, err
	}
	return delegatedPendingStakes, nil
}

func (delegatedPendingStakesList *DelegatedPendingStakesList) GetDelegatedPendingStakes(blockHeight uint64) (*delegated_pending_stakes.DelegatedPendingStakes, error) {
	key := strconv.FormatUint(blockHeight, 10)

	data, err := delegatedPendingStakesList.Get(key)
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*delegated_pending_stakes.DelegatedPendingStakes), nil
}

func NewDelegatedPendingStakesList(tx store_db_interface.StoreDBTransactionInterface) (delegatedPendingStakesList *DelegatedPendingStakesList) {

	hashmap := hash_map.CreateNewHashMap(tx, "pendingStakes", 0, false)

	delegatedPendingStakesList = &DelegatedPendingStakesList{
		HashMap: hashmap,
	}

	delegatedPendingStakesList.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return delegated_pending_stakes.NewDelegatedPendingStakes(key, index), nil
	}

	return
}
