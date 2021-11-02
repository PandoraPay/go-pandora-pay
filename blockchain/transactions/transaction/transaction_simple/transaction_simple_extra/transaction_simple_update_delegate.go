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
type TransactionSimpleExtraUpdateDelegate struct {
	TransactionSimpleExtraInterface
	DelegatedStakingClaimAmount uint64
	DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate
}

func (txExtra *TransactionSimpleExtraUpdateDelegate) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if err = txExtra.DelegatedStakingUpdate.Include(plainAcc); err != nil {
		return
	}

	if txExtra.DelegatedStakingClaimAmount > 0 {
		if err = plainAcc.AddUnclaimed(false, txExtra.DelegatedStakingClaimAmount); err != nil {
			return
		}
		if err = plainAcc.AddStakePendingStake(txExtra.DelegatedStakingClaimAmount, blockHeight); err != nil {
			return
		}
	}

	return
}

func (txExtra *TransactionSimpleExtraUpdateDelegate) Validate() (err error) {
	if err = txExtra.DelegatedStakingUpdate.Validate(); err != nil {
		return
	}

	if !txExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		if txExtra.DelegatedStakingClaimAmount == 0 {
			return errors.New("UpdateDelegateTx has no operation")
		}
	}

	return
}

func (txExtra *TransactionSimpleExtraUpdateDelegate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(txExtra.DelegatedStakingClaimAmount)
	txExtra.DelegatedStakingUpdate.Serialize(w)
}

func (txExtra *TransactionSimpleExtraUpdateDelegate) Deserialize(r *helpers.BufferReader) (err error) {
	if txExtra.DelegatedStakingClaimAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	txExtra.DelegatedStakingUpdate = &transaction_data.TransactionDataDelegatedStakingUpdate{}
	return txExtra.DelegatedStakingUpdate.Deserialize(r)
}
