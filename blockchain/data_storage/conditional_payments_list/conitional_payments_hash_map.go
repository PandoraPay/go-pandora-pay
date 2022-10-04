package conditional_payments_list

import (
	"pandora-pay/blockchain/data_storage/conditional_payments_list/conditional_payment"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type ConditionalPaymentsHashMap struct {
	*hash_map.HashMap[*conditional_payment.ConditionalPayment]
	BlockHeight uint64
}

func NewConditionalPaymentsHashMap(tx store_db_interface.StoreDBTransactionInterface, blockHeight uint64) (this *ConditionalPaymentsHashMap) {

	this = &ConditionalPaymentsHashMap{
		hash_map.CreateNewHashMap[*conditional_payment.ConditionalPayment](tx, "conditionalPayments_"+strconv.FormatUint(blockHeight, 10), 0, true),
		blockHeight,
	}

	this.HashMap.CreateObject = func(key []byte, index uint64) (*conditional_payment.ConditionalPayment, error) {
		return conditional_payment.NewConditionalPayment(key, index, blockHeight), nil
	}

	this.HashMap.StoredEvent = func(key []byte, committed *hash_map.CommittedMapElement[*conditional_payment.ConditionalPayment], index uint64) (err error) {
		if !this.Tx.IsWritable() {
			return
		}

		this.Tx.Put("conditionalPayments:all:"+string(key), []byte(strconv.FormatUint(committed.Element.BlockHeight, 10)))
		return
	}

	this.HashMap.DeletedEvent = func(key []byte) (err error) {
		if !this.Tx.IsWritable() {
			return
		}

		this.Tx.Delete("conditionalPayments:all:" + string(key))
		return
	}

	return
}
