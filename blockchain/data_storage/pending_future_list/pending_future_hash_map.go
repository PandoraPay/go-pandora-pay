package pending_future_list

import (
	"pandora-pay/blockchain/data_storage/pending_future_list/pending_future"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type PendingFutureHashMap struct {
	*hash_map.HashMap[*pending_future.PendingFuture]
	BlockHeight uint64
}

func NewPendingFutureHashMap(tx store_db_interface.StoreDBTransactionInterface, blockHeight uint64) (this *PendingFutureHashMap) {

	this = &PendingFutureHashMap{
		hash_map.CreateNewHashMap[*pending_future.PendingFuture](tx, "pendingFuture_"+strconv.FormatUint(blockHeight, 10), 0, true),
		blockHeight,
	}

	this.HashMap.CreateObject = func(key []byte, index uint64) (*pending_future.PendingFuture, error) {
		return pending_future.NewPendingFuture(key, index, blockHeight), nil
	}

	this.HashMap.StoredEvent = func(key []byte, committed *hash_map.CommittedMapElement[*pending_future.PendingFuture], index uint64) (err error) {
		if !this.Tx.IsWritable() {
			return
		}

		this.Tx.Put("pendingFuture:all:"+string(key), []byte(strconv.FormatUint(committed.Element.BlockHeight, 10)))
		return
	}

	this.HashMap.DeletedEvent = func(key []byte) (err error) {
		if !this.Tx.IsWritable() {
			return
		}

		this.Tx.Delete("pendingFuture:all:" + string(key))
		return
	}

	return
}
