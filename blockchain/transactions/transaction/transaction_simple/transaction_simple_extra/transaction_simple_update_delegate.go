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

	if plainAcc == nil {
		return errors.New("PlainAcc is null")
	}

	if tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		if !plainAcc.HasDelegatedStake() {
			if err = plainAcc.CreateDelegatedStake(0, tx.DelegatedStakingUpdate.DelegatedStakingNewPublicKey, tx.DelegatedStakingUpdate.DelegatedStakingNewFee); err != nil {
				return
			}
		} else {
			plainAcc.DelegatedStake.DelegatedStakePublicKey = tx.DelegatedStakingUpdate.DelegatedStakingNewPublicKey
			plainAcc.DelegatedStake.DelegatedStakeFee = tx.DelegatedStakingUpdate.DelegatedStakingNewFee
		}
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

func (tx *TransactionSimpleUpdateDelegate) Validate() error {
	if err := tx.DelegatedStakingUpdate.Validate(); err != nil {
		return err
	}

	if !tx.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
		if tx.DelegatedStakingClaimAmount == 0 {
			return errors.New("UpdateDelegateTx has no operation")
		}
	}

	return nil
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
