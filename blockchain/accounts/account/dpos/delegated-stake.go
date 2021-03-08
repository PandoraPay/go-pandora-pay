package dpos

import (
	"pandora-pay/config/stake"
	"pandora-pay/helpers"
)

type DelegatedStake struct {

	//public key for delegation
	DelegatedPublicKey [33]byte

	//confirmed stake
	StakeAvailable uint64

	//amount of the unstake
	UnstakeAvailable uint64

	//when unstake can be done
	UnstakeHeight uint64 // serialized only if UnstakeAvailable > 0

	//Pending stakes
	StakesPending []*DelegatedStakePending
}

func (dstake *DelegatedStake) AddUnstakeAvailable(sign bool, amount, blockHeight uint64) {
	if sign {
		helpers.SafeUint64Update(sign, &dstake.UnstakeAvailable, amount)
		dstake.UnstakeHeight = blockHeight + stake.GetUnstakeWindow(blockHeight)
	} else {
		if blockHeight < dstake.UnstakeHeight {
			panic("You can't withdraw now")
		}
		helpers.SafeUint64Update(sign, &dstake.UnstakeAvailable, amount)
	}
}

func (dstake *DelegatedStake) AddStakeAvailable(sign bool, amount uint64) {
	helpers.SafeUint64Update(sign, &dstake.StakeAvailable, amount)
}

func (dstake *DelegatedStake) AddStakePending(sign bool, amount, blockHeight uint64) {

	if amount == 0 {
		return
	}

	finalBlockHeight := blockHeight + stake.GetPendingStakeWindow(blockHeight)
	if sign {

		for _, stakePending := range dstake.StakesPending {
			if stakePending.ActivationHeight == finalBlockHeight {
				helpers.SafeUint64Add(&stakePending.PendingAmount, amount)
				return
			}
		}
		dstake.StakesPending = append(dstake.StakesPending, &DelegatedStakePending{
			ActivationHeight: finalBlockHeight,
			PendingAmount:    amount,
		})

	} else {

		for i, stakePending := range dstake.StakesPending {
			if stakePending.ActivationHeight == finalBlockHeight {
				helpers.SafeUint64Sub(&stakePending.PendingAmount, amount)
				if stakePending.PendingAmount == 0 {
					dstake.StakesPending = append(dstake.StakesPending[:i], dstake.StakesPending[i+1:]...)
				}
				return
			}
		}

		panic("Stake pending was not found!")
	}
}

func (dstake *DelegatedStake) Serialize(writer *helpers.BufferWriter) {

	writer.Write(dstake.DelegatedPublicKey[:])
	writer.WriteUvarint(dstake.StakeAvailable)
	writer.WriteUvarint(dstake.UnstakeAvailable)

	if dstake.UnstakeAvailable > 0 {
		writer.WriteUvarint(dstake.UnstakeHeight)
	}

	writer.WriteUvarint(uint64(len(dstake.StakesPending)))

	for _, stakePending := range dstake.StakesPending {
		stakePending.Serialize(writer)
	}

}

func (dstake *DelegatedStake) Deserialize(reader *helpers.BufferReader) {

	dstake.DelegatedPublicKey = reader.Read33()
	dstake.StakeAvailable = reader.ReadUvarint()
	dstake.UnstakeAvailable = reader.ReadUvarint()

	if dstake.UnstakeAvailable > 0 {
		dstake.UnstakeHeight = reader.ReadUvarint()
	}

	n := reader.ReadUvarint()
	for i := uint64(0); i < n; i++ {
		var delegatedStakePending = new(DelegatedStakePending)
		delegatedStakePending.Deserialize(reader)
		dstake.StakesPending = append(dstake.StakesPending, delegatedStakePending)
	}

	return
}

func (dstake *DelegatedStake) IsDelegatedStakeEmpty() bool {
	return dstake.StakeAvailable == 0 && dstake.UnstakeAvailable == 0 && len(dstake.StakesPending) == 0
}

func (dstake *DelegatedStake) RefreshDelegatedStake(blockHeight uint64) {

	for i := len(dstake.StakesPending) - 1; i >= 0; i-- {
		if dstake.StakesPending[i].ActivationHeight < blockHeight {
			dstake.StakeAvailable += dstake.StakesPending[i].PendingAmount
			dstake.StakesPending = append(dstake.StakesPending[:i], dstake.StakesPending[i+1:]...)
		}
	}

}

func (dstake *DelegatedStake) GetDelegatedStakeAvailable(blockHeight uint64) (result uint64) {

	result = dstake.StakeAvailable
	for i := range dstake.StakesPending {
		if dstake.StakesPending[i].ActivationHeight >= blockHeight {
			result += dstake.StakesPending[i].PendingAmount
		}
	}

	return
}
