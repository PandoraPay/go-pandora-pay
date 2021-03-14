package dpos

import (
	"pandora-pay/config/stake"
	"pandora-pay/helpers"
)

type DelegatedStake struct {

	//public key for delegation
	DelegatedPublicKey helpers.ByteString //33 bytes

	//confirmed stake
	StakeAvailable uint64

	//Pending stakes
	StakesPending []*DelegatedStakePending
}

func (dstake *DelegatedStake) AddStakeAvailable(sign bool, amount uint64) {
	helpers.SafeUint64Update(sign, &dstake.StakeAvailable, amount)
}

func (dstake *DelegatedStake) AddStakePendingStake(amount, blockHeight uint64) {
	if amount == 0 {
		return
	}
	finalBlockHeight := blockHeight + stake.GetPendingStakeWindow(blockHeight)

	for _, stakePending := range dstake.StakesPending {
		if stakePending.ActivationHeight == finalBlockHeight && stakePending.PendingType == true {
			helpers.SafeUint64Add(&stakePending.PendingAmount, amount)
			return
		}
	}
	dstake.StakesPending = append(dstake.StakesPending, &DelegatedStakePending{
		ActivationHeight: finalBlockHeight,
		PendingAmount:    amount,
		PendingType:      true,
	})

}

func (dstake *DelegatedStake) AddStakePendingUnstake(amount, blockHeight uint64) {
	if amount == 0 {
		return
	}
	finalBlockHeight := blockHeight + stake.GetUnstakeWindow(blockHeight)

	for _, stakePending := range dstake.StakesPending {
		if stakePending.ActivationHeight == finalBlockHeight && stakePending.PendingType == false {
			helpers.SafeUint64Add(&stakePending.PendingAmount, amount)
			stakePending.ActivationHeight = finalBlockHeight
			return
		}
	}
	dstake.StakesPending = append(dstake.StakesPending, &DelegatedStakePending{
		ActivationHeight: finalBlockHeight,
		PendingAmount:    amount,
		PendingType:      false,
	})

}

func (dstake *DelegatedStake) Serialize(writer *helpers.BufferWriter) {

	writer.Write(dstake.DelegatedPublicKey)
	writer.WriteUvarint(dstake.StakeAvailable)

	writer.WriteUvarint(uint64(len(dstake.StakesPending)))
	for _, stakePending := range dstake.StakesPending {
		stakePending.Serialize(writer)
	}

}

func (dstake *DelegatedStake) Deserialize(reader *helpers.BufferReader) {

	dstake.DelegatedPublicKey = reader.ReadBytes(33)
	dstake.StakeAvailable = reader.ReadUvarint()

	n := reader.ReadUvarint()
	dstake.StakesPending = make([]*DelegatedStakePending, n)
	for i := uint64(0); i < n; i++ {
		delegatedStakePending := new(DelegatedStakePending)
		delegatedStakePending.Deserialize(reader)
		dstake.StakesPending[i] = delegatedStakePending
	}

	return
}

func (dstake *DelegatedStake) IsDelegatedStakeEmpty() bool {
	return dstake.StakeAvailable == 0 && len(dstake.StakesPending) == 0
}

func (dstake *DelegatedStake) GetDelegatedStakeAvailable(blockHeight uint64) (result uint64) {

	result = dstake.StakeAvailable
	for i := range dstake.StakesPending {
		if dstake.StakesPending[i].ActivationHeight <= blockHeight && dstake.StakesPending[i].PendingType {
			helpers.SafeUint64Add(&result, dstake.StakesPending[i].PendingAmount)
		}
	}

	return
}

func (dstake *DelegatedStake) GetDelegatedUnstakeAvailable(blockHeight uint64) (result uint64) {

	for i := range dstake.StakesPending {
		if dstake.StakesPending[i].ActivationHeight <= blockHeight && !dstake.StakesPending[i].PendingType {
			helpers.SafeUint64Add(&result, dstake.StakesPending[i].PendingAmount)
		}
	}

	return
}
