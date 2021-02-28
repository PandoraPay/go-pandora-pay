package dpos

import (
	"bytes"
	"encoding/binary"
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

func (delegatedStake *DelegatedStake) Serialize(serialized *bytes.Buffer, temp []byte) {

	serialized.Write(delegatedStake.DelegatedPublicKey[:])

	n := binary.PutUvarint(temp, delegatedStake.StakeAvailable)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, delegatedStake.UnstakeAmount)
	serialized.Write(temp[:n])

	if delegatedStake.UnstakeAmount > 0 {
		n = binary.PutUvarint(temp, delegatedStake.UnstakeHeight)
		serialized.Write(temp[:n])
	}

	n = binary.PutUvarint(temp, uint64(len(delegatedStake.StakesPending)))
	serialized.Write(temp[:n])

	for _, stakePending := range delegatedStake.StakesPending {
		stakePending.Serialize(serialized, temp)
	}

}

func (delegatedStake *DelegatedStake) Deserialize(reader *helpers.BufferReader) (err error) {

	var data []byte
	if data, err = reader.ReadBytes(33); err != nil {
		return
	}
	delegatedStake.DelegatedPublicKey = *helpers.Byte33(data)

	if delegatedStake.StakeAvailable, err = reader.ReadUvarint(); err != nil {
		return
	}

	if delegatedStake.UnstakeAmount, err = reader.ReadUvarint(); err != nil {
		return
	}

	if delegatedStake.UnstakeAmount > 0 {
		if delegatedStake.UnstakeHeight, err = reader.ReadUvarint(); err != nil {
			return
		}
	}

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}

	for i := uint64(0); i < n; i++ {
		var delegatedStakePending = new(DelegatedStakePending)
		if err = delegatedStakePending.Deserialize(reader); err != nil {
			return
		}

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
	for _, stakePending := range delegatedStake.StakesPending {
		if stakePending.StakePendingHeight >= blockHeight {
			result += stakePending.StakePending
		}
	}

	return
}
