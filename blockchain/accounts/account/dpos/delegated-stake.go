package dpos

import (
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-"`
	DelegatedPublicKeyHash        helpers.HexBytes         `json:"delegatedPublicKeyHash"` //public key for delegation  20 bytes
	StakeAvailable                uint64                   `json:"stakeAvailable"`         //confirmed stake
	StakesPending                 []*DelegatedStakePending `json:"stakesPending"`          //Pending stakes
}

func (dstake *DelegatedStake) AddStakeAvailable(sign bool, amount uint64) error {
	return helpers.SafeUint64Update(sign, &dstake.StakeAvailable, amount)
}

func (dstake *DelegatedStake) AddStakePendingStake(amount, blockHeight uint64) error {

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

func (dstake *DelegatedStake) Serialize(writer *helpers.BufferWriter) {

	writer.Write(dstake.DelegatedPublicKeyHash)
	writer.WriteUvarint(dstake.StakeAvailable)

	writer.WriteUvarint(uint64(len(dstake.StakesPending)))
	for _, stakePending := range dstake.StakesPending {
		stakePending.Serialize(writer)
	}

}

func (dstake *DelegatedStake) Deserialize(reader *helpers.BufferReader) (err error) {

	if dstake.DelegatedPublicKeyHash, err = reader.ReadBytes(cryptography.PublicKeyHashHashSize); err != nil {
		return
	}
	if dstake.StakeAvailable, err = reader.ReadUvarint(); err != nil {
		return
	}

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	dstake.StakesPending = make([]*DelegatedStakePending, n)
	for i := uint64(0); i < n; i++ {
		delegatedStakePending := new(DelegatedStakePending)
		if err = delegatedStakePending.Deserialize(reader); err != nil {
			return
		}
		dstake.StakesPending[i] = delegatedStakePending
	}

	return
}

func (dstake *DelegatedStake) IsDelegatedStakeEmpty() bool {
	return dstake.StakeAvailable == 0 && len(dstake.StakesPending) == 0
}

func (dstake *DelegatedStake) GetDelegatedStakeAvailable() uint64 {
	return dstake.StakeAvailable
}

func (dstake *DelegatedStake) ComputeDelegatedStakeAvailable(blockHeight uint64) (uint64, error) {
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
