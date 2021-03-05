package dpos

import (
	"pandora-pay/config"
	"pandora-pay/helpers"
)

type DelegatedStake struct {

	//public key for delegation
	DelegatedPublicKey [33]byte

	//confirmed stake
	StakeAvailable uint64

	//amount of the unstake
	UnstakeAmount uint64

	//when unstake can be done
	UnstakeHeight uint64 // serialized only if UnstakeAmount > 0

	//Pending stakes
	StakesPending []*DelegatedStakePending
}

func (delegatedStake *DelegatedStake) AddDelegatedUnstake(sign bool, amount, blockHeight uint64) {
	helpers.SafeUint64Update(sign, &delegatedStake.UnstakeAmount, amount)
	delegatedStake.UnstakeHeight = blockHeight + config.UNSTAKE_BLOCK_WINDOW
}

func (delegatedStake *DelegatedStake) AddDelegatedStake(sign bool, amount uint64) {
	helpers.SafeUint64Update(sign, &delegatedStake.StakeAvailable, amount)
}

func (delegatedStake *DelegatedStake) Serialize(writer *helpers.BufferWriter) {

	writer.Write(delegatedStake.DelegatedPublicKey[:])
	writer.WriteUvarint(delegatedStake.StakeAvailable)
	writer.WriteUvarint(delegatedStake.UnstakeAmount)

	if delegatedStake.UnstakeAmount > 0 {
		writer.WriteUvarint(delegatedStake.UnstakeHeight)
	}

	writer.WriteUvarint(uint64(len(delegatedStake.StakesPending)))

	for _, stakePending := range delegatedStake.StakesPending {
		stakePending.Serialize(writer)
	}

}

func (delegatedStake *DelegatedStake) Deserialize(reader *helpers.BufferReader) {

	delegatedStake.DelegatedPublicKey = reader.Read33()
	delegatedStake.StakeAvailable = reader.ReadUvarint()
	delegatedStake.UnstakeAmount = reader.ReadUvarint()

	if delegatedStake.UnstakeAmount > 0 {
		delegatedStake.UnstakeHeight = reader.ReadUvarint()
	}

	n := reader.ReadUvarint()
	for i := uint64(0); i < n; i++ {
		var delegatedStakePending = new(DelegatedStakePending)
		delegatedStakePending.Deserialize(reader)
		delegatedStake.StakesPending = append(delegatedStake.StakesPending, delegatedStakePending)
	}

	return
}

func (delegatedStake *DelegatedStake) IsDelegatedStakeEmpty() bool {
	return delegatedStake.StakeAvailable == 0 && delegatedStake.UnstakeAmount == 0 && len(delegatedStake.StakesPending) == 0
}

func (delegatedStake *DelegatedStake) RefreshDelegatedStake(blockHeight uint64) {

	for i := len(delegatedStake.StakesPending) - 1; i >= 0; i-- {
		if delegatedStake.StakesPending[i].StakePendingHeight < blockHeight {
			delegatedStake.StakeAvailable += delegatedStake.StakesPending[i].StakePending
			delegatedStake.StakesPending = append(delegatedStake.StakesPending[:i], delegatedStake.StakesPending[i+1:]...)
		}
	}

}

func (delegatedStake *DelegatedStake) GetDelegatedStakeAvailable(blockHeight uint64) (result uint64) {

	result = delegatedStake.StakeAvailable
	for i := range delegatedStake.StakesPending {
		if delegatedStake.StakesPending[i].StakePendingHeight >= blockHeight {
			result += delegatedStake.StakesPending[i].StakePending
		}
	}

	return
}
