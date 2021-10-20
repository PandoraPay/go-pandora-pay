package transaction_data

import (
	"errors"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionDataDelegatedStakingUpdate struct {
	DelegatedStakingHasNewInfo   bool
	DelegatedStakingNewPublicKey helpers.HexBytes //20 byte
	DelegatedStakingNewFee       uint64
}

func (data *TransactionDataDelegatedStakingUpdate) Validate() error {
	if data.DelegatedStakingHasNewInfo {
		if len(data.DelegatedStakingNewPublicKey) != cryptography.PublicKeySize {
			return errors.New("New Public Key Hash length is invalid")
		}
		if data.DelegatedStakingNewFee > config_stake.DELEGATING_STAKING_FEES_MAX_VALUE {
			return errors.New("Invalid NewDelegatedStakingNewFee")
		}
	} else {
		if len(data.DelegatedStakingNewPublicKey) != 0 {
			return errors.New("New Public Key Hash length is invalid")
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
