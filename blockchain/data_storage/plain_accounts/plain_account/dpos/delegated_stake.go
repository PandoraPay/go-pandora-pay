package dpos

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Version                       DelegatedStakeVersion `json:"version" msgpack:"version"`
	DelegatedStakePublicKey       []byte                `json:"delegatedStakePublicKey,omitempty" msgpack:"delegatedStakePublicKey,omitempty"` //public key for delegation  20 bytes
	DelegatedStakeFee             uint64                `json:"delegatedStakeFee,omitempty" msgpack:"delegatedStakeFee,omitempty"`
	StakeAvailable                uint64                `json:"stakeAvailable,omitempty" msgpack:"stakeAvailable,omitempty"` //confirmed stake
}

func (dstake *DelegatedStake) IsDeletable() bool {
	return dstake.Version == NO_STAKING
}

func (dstake *DelegatedStake) Validate() error {
	switch dstake.Version {
	case NO_STAKING:
	case STAKING:
	default:
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (dstake *DelegatedStake) HasDelegatedStake() bool {
	return dstake.Version == STAKING
}

func (dstake *DelegatedStake) AddStakeAvailable(sign bool, amount uint64) error {
	if !dstake.HasDelegatedStake() {
		return errors.New("plainAccount.HasDelegatedStake is false")
	}

	return helpers.SafeUint64Update(sign, &dstake.StakeAvailable, amount)
}

func (dstake *DelegatedStake) CreateDelegatedStake(amount uint64, delegatedStakePublicKey []byte, delegatedStakeFee uint64) error {
	if dstake.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}

	if delegatedStakePublicKey == nil || len(delegatedStakePublicKey) != cryptography.PublicKeySize {
		return errors.New("delegatedStakePublicKey is Invalid")
	}

	dstake.Version = STAKING
	dstake.StakeAvailable = amount
	dstake.DelegatedStakePublicKey = delegatedStakePublicKey
	dstake.DelegatedStakeFee = delegatedStakeFee

	return nil
}

func (dstake *DelegatedStake) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(uint64(dstake.Version))
	if dstake.Version == STAKING {
		w.Write(dstake.DelegatedStakePublicKey)
		w.WriteUvarint(dstake.StakeAvailable)
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
		if dstake.DelegatedStakePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if dstake.StakeAvailable, err = r.ReadUvarint(); err != nil {
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

func (dstake *DelegatedStake) GetDelegatedStakeAvailable() uint64 {
	if !dstake.HasDelegatedStake() {
		return 0
	}
	return dstake.StakeAvailable
}
