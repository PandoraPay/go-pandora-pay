package dpos

import (
	"errors"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Version                       DelegatedStakeVersion `json:"version" msgpack:"version"`
	DelegatedStakeNonce           uint64                `json:"delegatedStakeNonce" msgpack:"delegatedStakeNonce"`
	DelegatedStakePublicKey       []byte                `json:"delegatedStakePublicKey,omitempty" msgpack:"delegatedStakePublicKey,omitempty"` //public key for delegation  20 bytes
	DelegatedStakeFee             uint64                `json:"delegatedStakeFee,omitempty" msgpack:"delegatedStakeFee,omitempty"`
}

func (dstake *DelegatedStake) IsDeletable() bool {
	return dstake.Version == NO_STAKING
}

func (dstake *DelegatedStake) Validate() error {
	switch dstake.Version {
	case NO_STAKING:
		if len(dstake.DelegatedStakePublicKey) > 0 || dstake.DelegatedStakeFee > 0 || dstake.DelegatedStakeNonce > 0 {
			return errors.New("DelegatedStake should have empty data")
		}
	case STAKING:
		if len(dstake.DelegatedStakePublicKey) != cryptography.PublicKeySize {
			return errors.New("DelegatedStakePublicKey length is invalid")
		}
		if dstake.DelegatedStakeFee > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
			return errors.New("dstake.DelegatedStakeFee is invalid")
		}
	default:
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (dstake *DelegatedStake) HasDelegatedStake() bool {
	return dstake.Version == STAKING
}

func (dstake *DelegatedStake) CreateDelegatedStake(delegatedStakeNonce uint64, delegatedStakePublicKey []byte, delegatedStakeFee uint64) error {

	if len(delegatedStakePublicKey) != cryptography.PublicKeySize {
		return errors.New("delegatedStakePublicKey is Invalid")
	}
	if delegatedStakeFee > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
		return errors.New("delegatedStakeFee is invalid")
	}

	if !dstake.HasDelegatedStake() {
		dstake.Version = STAKING
	}

	dstake.DelegatedStakeNonce = delegatedStakeNonce
	dstake.DelegatedStakePublicKey = delegatedStakePublicKey
	dstake.DelegatedStakeFee = delegatedStakeFee

	return nil
}

func (dstake *DelegatedStake) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(uint64(dstake.Version))
	if dstake.Version == STAKING {
		w.WriteUvarint(dstake.DelegatedStakeNonce)
		w.Write(dstake.DelegatedStakePublicKey)
		w.WriteUvarint(dstake.DelegatedStakeFee)
	}
}

func (dstake *DelegatedStake) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	dstake.Version = DelegatedStakeVersion(n)

	switch dstake.Version {
	case NO_STAKING:
	case STAKING:
		if dstake.DelegatedStakeNonce, err = r.ReadUvarint(); err != nil {
			return
		}
		if dstake.DelegatedStakePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if dstake.DelegatedStakeFee, err = r.ReadUvarint(); err != nil {
			return
		}
	default:
		return errors.New("Invalid DelegatedStake version")
	}

	return
}
