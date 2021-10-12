package transaction_zether_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherDelegateStake struct {
	TransactionZetherExtraInterface
	DelegatePublicKey []byte
}

func (tx *TransactionZetherDelegateStake) IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, dataStorage *data_storage.DataStorage) error {
	return nil
}

func (tx *TransactionZetherDelegateStake) Validate() error {
	if len(tx.DelegatePublicKey) != cryptography.PublicKeySize {
		return errors.New("DelegatePublicKey is invalid")
	}
	return nil
}

func (tx *TransactionZetherDelegateStake) Serialize(w *helpers.BufferWriter) {
	w.Write(tx.DelegatePublicKey)
}

func (tx *TransactionZetherDelegateStake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	return
}
