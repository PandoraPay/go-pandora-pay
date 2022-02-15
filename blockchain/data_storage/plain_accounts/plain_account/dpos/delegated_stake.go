package dpos

import (
	"errors"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Version                       DelegatedStakeVersion    `json:"version" msgpack:"version"`
	DelegatedStakePublicKey       []byte                   `json:"delegatedStakePublicKey,omitempty" msgpack:"delegatedStakePublicKey,omitempty"` //public key for delegation  20 bytes
	DelegatedStakeFee             uint64                   `json:"delegatedStakeFee,omitempty" msgpack:"delegatedStakeFee,omitempty"`
	StakeAvailable                uint64                   `json:"stakeAvailable,omitempty" msgpack:"stakeAvailable,omitempty"` //confirmed stake
	StakesPending                 []*DelegatedStakePending `json:"stakesPending,omitempty" msgpack:"stakesPending,omitempty"`   //Pending stakes
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

func (dstake *DelegatedStake) AddStakePendingStake(amount, blockHeight uint64) error {

	if !dstake.HasDelegatedStake() {
		return errors.New("plainAccount.HasDelegatedStake is false")
	}

	if amount == 0 {
		return nil
	}
	finalBlockHeight := blockHeight + config_stake.GetPendingStakeWindow(blockHeight)

	for _, stakePending := range dstake.StakesPending {
		if stakePending.ActivationHeight == finalBlockHeight && stakePending.PendingType == DelegatedStakePendingStake {
			return helpers.SafeUint64Add(&stakePending.PendingAmount, amount)
		}
	}

	dstake.StakesPending = append(dstake.StakesPending, &DelegatedStakePending{
		ActivationHeight: finalBlockHeight,
		PendingAmount:    amount,
		PendingType:      DelegatedStakePendingStake,
	})

	return nil
}

func (dstake *DelegatedStake) AddStakePendingUnstake(amount, blockHeight uint64) error {
	if !dstake.HasDelegatedStake() {
		return errors.New("plainAccount.HasDelegatedStake is false")
	}

	if amount == 0 {
		return nil
	}
	finalBlockHeight := blockHeight + config_stake.GetUnstakeWindow(blockHeight)

	for _, stakePending := range dstake.StakesPending {
		if stakePending.ActivationHeight == finalBlockHeight && stakePending.PendingType == DelegatedStakePendingUnstake {
			stakePending.ActivationHeight = finalBlockHeight
			return helpers.SafeUint64Add(&stakePending.PendingAmount, amount)
		}
	}
	dstake.StakesPending = append(dstake.StakesPending, &DelegatedStakePending{
		ActivationHeight: finalBlockHeight,
		PendingAmount:    amount,
		PendingType:      DelegatedStakePendingUnstake,
	})
	return nil
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
	dstake.StakesPending = []*DelegatedStakePending{}
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

		w.WriteUvarint(uint64(len(dstake.StakesPending)))
		for _, stakePending := range dstake.StakesPending {
			stakePending.Serialize(w)
		}
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

		var n uint64
		if n, err = r.ReadUvarint(); err != nil {
			return
		}

		dstake.StakesPending = make([]*DelegatedStakePending, n)
		for i := uint64(0); i < n; i++ {
			delegatedStakePending := new(DelegatedStakePending)
			if err = delegatedStakePending.Deserialize(r); err != nil {
				return
			}
			dstake.StakesPending[i] = delegatedStakePending
		}
	default:
		return errors.New("Invalid DelegatedStake version")
	}

	return
}

func (dstake *DelegatedStake) isDelegatedStakeEmpty() bool {
	return dstake.StakeAvailable == 0 && len(dstake.StakesPending) == 0
}

func (dstake *DelegatedStake) GetDelegatedStakeAvailable() uint64 {
	if !dstake.HasDelegatedStake() {
		return 0
	}
	return dstake.StakeAvailable
}

func (dstake *DelegatedStake) ComputeDelegatedStakeAvailable(blockHeight uint64) (uint64, error) {
	if !dstake.HasDelegatedStake() {
		return 0, nil
	}

	result := dstake.StakeAvailable
	for i := range dstake.StakesPending {
		if dstake.StakesPending[i].ActivationHeight <= blockHeight && dstake.StakesPending[i].PendingType == DelegatedStakePendingStake {
			if err := helpers.SafeUint64Add(&result, dstake.StakesPending[i].PendingAmount); err != nil {
				return 0, err
			}
		}
	}
	return result, nil
}

func (dstake *DelegatedStake) ComputeDelegatedUnstakePending() (uint64, error) {
	if !dstake.HasDelegatedStake() {
		return 0, nil
	}

	result := uint64(0)
	for i := range dstake.StakesPending {
		if dstake.StakesPending[i].PendingType == DelegatedStakePendingUnstake {
			if err := helpers.SafeUint64Add(&result, dstake.StakesPending[i].PendingAmount); err != nil {
				return 0, err
			}
		}
	}
	return result, nil
}

func (dstake *DelegatedStake) RefreshDelegatedStake(blockHeight uint64) (uint64, error) {

	unclaimed := uint64(0)
	var err error
	for i := len(dstake.StakesPending) - 1; i >= 0; i-- {
		stakePending := dstake.StakesPending[i]
		if stakePending.ActivationHeight <= blockHeight {

			if stakePending.PendingType == DelegatedStakePendingStake {
				if err = dstake.AddStakeAvailable(true, stakePending.PendingAmount); err != nil {
					return 0, err
				}
			} else {
				if err = helpers.SafeUint64Add(&unclaimed, stakePending.PendingAmount); err != nil {
					return 0, err
				}
			}
			dstake.StakesPending = append(dstake.StakesPending[:i], dstake.StakesPending[i+1:]...)
		}
	}

	if dstake.isDelegatedStakeEmpty() {
		dstake.Version = NO_STAKING
		dstake.DelegatedStakePublicKey = nil
		dstake.DelegatedStakeFee = 0
		dstake.StakeAvailable = 0
		dstake.StakesPending = []*DelegatedStakePending{}
	}

	return unclaimed, nil
}
