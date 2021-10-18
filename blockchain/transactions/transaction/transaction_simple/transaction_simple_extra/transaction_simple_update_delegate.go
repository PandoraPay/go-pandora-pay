package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

/**
Substracting DelegatedStakingClaimAmount from the Unclaimed
Creating a Stake Pending with DelegatedStakingClaimAmount
*/
type TransactionSimpleUpdateDelegate struct {
	TransactionSimpleExtraInterface
	DelegatedStakingClaimAmount  uint64
	DelegatedStakingHasNewInfo   bool
	DelegatedStakingNewPublicKey helpers.HexBytes //20 byte
	DelegatedStakingNewFee       uint64
}

func (tx *TransactionSimpleUpdateDelegate) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if plainAcc == nil {
		return errors.New("PlainAcc is null")
	}

	if tx.DelegatedStakingHasNewInfo {
		if !plainAcc.HasDelegatedStake() {
			if err = plainAcc.CreateDelegatedStake(0, tx.DelegatedStakingNewPublicKey, tx.DelegatedStakingNewFee); err != nil {
				return
			}
		} else {
			plainAcc.DelegatedStake.DelegatedStakePublicKey = tx.DelegatedStakingNewPublicKey
			plainAcc.DelegatedStake.DelegatedStakeFee = tx.DelegatedStakingNewFee
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
	if tx.DelegatedStakingHasNewInfo {
		if len(tx.DelegatedStakingNewPublicKey) != cryptography.PublicKeySize {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.DelegatedStakingNewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
			return errors.New("Invalid NewDelegatedStakingNewFee")
		}
	} else {
		if len(tx.DelegatedStakingNewPublicKey) != 0 {
			return errors.New("New Public Key Hash length is invalid")
		}
		if tx.DelegatedStakingNewFee != 0 {
			return errors.New("Invalid NewDelegatedStakingNewFee")
		}
		if tx.DelegatedStakingClaimAmount == 0 {
			return errors.New("UpdateDelegateTx has no operation")
		}
	}
	return nil
}

func (tx *TransactionSimpleUpdateDelegate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(tx.DelegatedStakingClaimAmount)
	w.WriteBool(tx.DelegatedStakingHasNewInfo)
	if tx.DelegatedStakingHasNewInfo {
		w.Write(tx.DelegatedStakingNewPublicKey)
		w.WriteUvarint(tx.DelegatedStakingNewFee)
	}
}

func (tx *TransactionSimpleUpdateDelegate) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.DelegatedStakingClaimAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if tx.DelegatedStakingHasNewInfo, err = r.ReadBool(); err != nil {
		return
	}

	if tx.DelegatedStakingHasNewInfo {
		if tx.DelegatedStakingNewPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if tx.DelegatedStakingNewFee, err = r.ReadUvarint(); err != nil {
			return
		}
	}
	return
}
