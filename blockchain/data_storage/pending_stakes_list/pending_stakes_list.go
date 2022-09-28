package pending_stakes_list

import (
	"pandora-pay/blockchain/data_storage/pending_stakes_list/pending_stakes"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type PendingStakesList struct {
	*hash_map.HashMap[*pending_stakes.PendingStakes]
}

func (this *PendingStakesList) CreateNewPendingStakes(blockHeight uint64) (*pending_stakes.PendingStakes, error) {
	key := strconv.FormatUint(blockHeight, 10)

	pendingStakes := pending_stakes.NewPendingStakes([]byte(key), 0) //index will be set by update
	pendingStakes.Height = blockHeight

	if err := this.Create(key, pendingStakes); err != nil {
		return nil, err
	}
	return pendingStakes, nil
}

func (this *PendingStakesList) GetPendingStakes(blockHeight uint64) (*pending_stakes.PendingStakes, error) {
	return this.Get(strconv.FormatUint(blockHeight, 10))
}

func NewPendingStakesList(tx store_db_interface.StoreDBTransactionInterface) (this *PendingStakesList) {

	this = &PendingStakesList{
		hash_map.CreateNewHashMap[*pending_stakes.PendingStakes](tx, "pendingStakes", 0, false),
	}

	this.HashMap.CreateObject = func(key []byte, index uint64) (*pending_stakes.PendingStakes, error) {
		return pending_stakes.NewPendingStakes(key, index), nil
	}

	return
}
