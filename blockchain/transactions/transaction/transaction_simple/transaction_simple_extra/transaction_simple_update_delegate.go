package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

/**
Substracting UpdateStakingAmount from the Claimable
Creating a Stake Pending with UpdateStakingAmount
*/
type TransactionSimpleUpdateDelegate struct {
	TransactionSimpleExtraInterface
	UpdateStakingAmount uint64
	HasNewDelegatedInfo bool
	NewPublicKey        helpers.HexBytes //20 byte
	NewFee              uint64
}

func (tx *TransactionSimpleUpdateDelegate) IncludeTransactionVin0(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if len(txRegistrations.Registrations) > 0 {
		return errors.New("txRegistrations.Registrations length should be zero")
	}

	if plainAcc == nil {
		return errors.New("PlainAcc is null")
	}

	if !plainAcc.HasDelegatedStake() {
		if err = plainAcc.CreateDelegatedStake(0, tx.NewPublicKey, tx.NewFee); err != nil {
			return
		}
	} else {
		plainAcc.DelegatedStake.DelegatedStakePublicKey = tx.NewPublicKey
		plainAcc.DelegatedStake.DelegatedStakeFee = tx.NewFee
	}

	if tx.UpdateStakingAmount > 0 {
		if err = plainAcc.AddClaimable(false, tx.UpdateStakingAmount); err != nil {
			return
		}
		if err = plainAcc.DelegatedStake.AddStakePendingStake(tx.UpdateStakingAmount, blockHeight); err != nil {
			return
		}
	}

	return
}

func (tx *TransactionSimpleUpdateDelegate) Validate() error {
	if tx.HasNewDelegatedInfo {
		if len(tx.NewPublicKey) != cryptography.PublicKeySize {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.NewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
			return errors.New("Invalid NewFee")
		}
	} else {
		if len(tx.NewPublicKey) != 0 {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.NewFee != 0 {
			return errors.New("Invalid NewFee")
		}
		if tx.UpdateStakingAmount == 0 {
			return errors.New("UpdateDelegateTx has no operation")
		}
	}
	return nil
}

func (tx *TransactionSimpleUpdateDelegate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(tx.UpdateStakingAmount)
	w.WriteBool(tx.HasNewDelegatedInfo)
	if tx.HasNewDelegatedInfo {
		w.Write(tx.NewPublicKey)
		w.WriteUvarint(tx.NewFee)
	}
}

func (tx *TransactionSimpleUpdateDelegate) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.UpdateStakingAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if tx.HasNewDelegatedInfo, err = r.ReadBool(); err != nil {
		return
	}

	if tx.HasNewDelegatedInfo {
		if tx.NewPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if tx.NewFee, err = r.ReadUvarint(); err != nil {
			return
		}
	}
	return
}
