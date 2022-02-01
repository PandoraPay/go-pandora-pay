package transaction_data

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionDataDelegatedStakingUpdate struct {
	DelegatedStakingHasNewInfo   bool             `json:"delegatedStakingHasNewInfo" msgpack:"delegatedStakingHasNewInfo"`
	DelegatedStakingNewPublicKey helpers.HexBytes `json:"delegatedStakingNewPublicKey" msgpack:"delegatedStakingNewPublicKey"` //20 byte
	DelegatedStakingNewFee       uint64           `json:"delegatedStakingNewFee" msgpack:"delegatedStakingNewFee"`
}

func (data *TransactionDataDelegatedStakingUpdate) Include(plainAcc *plain_account.PlainAccount) (err error) {

	if plainAcc == nil {
		return errors.New("PlainAcc is null")
	}

	if data.DelegatedStakingHasNewInfo {
		if !plainAcc.DelegatedStake.HasDelegatedStake() {
			if err = plainAcc.DelegatedStake.CreateDelegatedStake(0, data.DelegatedStakingNewPublicKey, data.DelegatedStakingNewFee); err != nil {
				return
			}
		} else {
			plainAcc.DelegatedStake.DelegatedStakePublicKey = data.DelegatedStakingNewPublicKey
			plainAcc.DelegatedStake.DelegatedStakeFee = data.DelegatedStakingNewFee
		}
	}
	return
}

func (data *TransactionDataDelegatedStakingUpdate) Validate() error {
	if data.DelegatedStakingHasNewInfo {
		if len(data.DelegatedStakingNewPublicKey) != cryptography.PublicKeySize {
			return fmt.Errorf("New Public Key Hash length is invalid. It should be %d", cryptography.PublicKeySize)
		}
		if data.DelegatedStakingNewFee > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
			return errors.New("Invalid NewDelegatedStakingNewFee")
		}
	} else {
		if len(data.DelegatedStakingNewPublicKey) != 0 {
			return errors.New("New Public Key Hash length is invalid. It should have been zero")
		}
		if data.DelegatedStakingNewFee != 0 {
			return errors.New("Invalid NewDelegatedStakingNewFee")
		}
	}
	return nil
}

func (data *TransactionDataDelegatedStakingUpdate) Serialize(w *helpers.BufferWriter) {
	w.WriteBool(data.DelegatedStakingHasNewInfo)
	if data.DelegatedStakingHasNewInfo {
		w.Write(data.DelegatedStakingNewPublicKey)
		w.WriteUvarint(data.DelegatedStakingNewFee)
	}
}

func (data *TransactionDataDelegatedStakingUpdate) Deserialize(r *helpers.BufferReader) (err error) {
	if data.DelegatedStakingHasNewInfo, err = r.ReadBool(); err != nil {
		return
	}

	if data.DelegatedStakingHasNewInfo {
		if data.DelegatedStakingNewPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if data.DelegatedStakingNewFee, err = r.ReadUvarint(); err != nil {
			return
		}
	}

	return
}
