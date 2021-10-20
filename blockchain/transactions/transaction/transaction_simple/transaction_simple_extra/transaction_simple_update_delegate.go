package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

/**
Substracting DelegatedStakingClaimAmount from the Unclaimed
Creating a Stake Pending with DelegatedStakingClaimAmount
*/
type TransactionSimpleUpdateDelegate struct {
	TransactionSimpleExtraInterface
	DelegatedStakingClaimAmount uint64
	DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate
}

func (tx *TransactionSimpleUpdateDelegate) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if err = tx.DelegatedStakingUpdate.Include(plainAcc); err != nil {
		return
	}

	if tx.DelegatedStakingClaimAmount > 0 {
		if err = plainAcc.AddUnclaimed(false, tx.DelegatedStakingClaimAmount); err != nil {
			return
		}
		if err = plainAcc.DelegatedStake.AddStakePendingStake(tx.DelegatedStakingClaimAmount, blockHeight); err != nil {
			return
		}
	}

	return
}

func (tx *TransactionSimpleUpdateDelegate) Validate() (err error) {
	if err = tx.DelegatedStakingUpdate.Validate(); err != nil {
		return
	}

	if !tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		if tx.DelegatedStakingClaimAmount == 0 {
			return errors.New("UpdateDelegateTx has no operation")
		}
	}

	return
}

func (tx *TransactionSimpleUpdateDelegate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(tx.DelegatedStakingClaimAmount)
	tx.DelegatedStakingUpdate.Serialize(w)
}

func (tx *TransactionSimpleUpdateDelegate) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatedStakingClaimAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	tx.DelegatedStakingUpdate = &transaction_data.TransactionDataDelegatedStakingUpdate{}
	return tx.DelegatedStakingUpdate.Deserialize(r)
}
