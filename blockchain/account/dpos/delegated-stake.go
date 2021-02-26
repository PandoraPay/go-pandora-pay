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
	UnstakeHeight uint64

	//Pending stakes
	StakesPending []*DelegatedStakePending
}

func (delegatedStake *DelegatedStake) Serialize(serialized *bytes.Buffer, temp []byte) {

	serialized.Write(delegatedStake.DelegatedPublicKey[:])

	n := binary.PutUvarint(temp, delegatedStake.StakeAvailable)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, delegatedStake.UnstakeAmount)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, delegatedStake.UnstakeHeight)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, uint64(len(delegatedStake.StakesPending)))
	serialized.Write(temp[:n])

	for _, stakePending := range delegatedStake.StakesPending {
		stakePending.Serialize(serialized, temp)
	}

}

func (delegatedStake *DelegatedStake) Deserialize(buf []byte) (out []byte, err error) {

	var data []byte
	if data, buf, err = helpers.DeserializeBuffer(buf, 33); err != nil {
		return
	}
	delegatedStake.DelegatedPublicKey = *helpers.Byte33(data)

	if delegatedStake.StakeAvailable, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	if delegatedStake.UnstakeAmount, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	if delegatedStake.UnstakeHeight, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	var n uint64
	if n, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	for i := uint64(0); i < n; i++ {
		var delegatedStakePending = new(DelegatedStakePending)
		if buf, err = delegatedStakePending.Deserialize(buf); err != nil {
			return
		}

		delegatedStake.StakesPending = append(delegatedStake.StakesPending, delegatedStakePending)
	}

	out = buf
	return
}

func (delegatedStake *DelegatedStake) IsDelegatedStakeEmpty() bool {
	return delegatedStake.StakeAvailable == 0 && delegatedStake.UnstakeAmount == 0 && len(delegatedStake.StakesPending) == 0
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
