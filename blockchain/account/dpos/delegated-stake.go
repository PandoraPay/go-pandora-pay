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
	StakeConfirmed uint64

	//when unstake can be done
	UnstakeHeight uint64

	//Pending stakes
	StakesPending []*DelegatedStakePending
}

func (delegatedStake *DelegatedStake) Serialize(serialized *bytes.Buffer, temp []byte) {

	serialized.Write(delegatedStake.DelegatedPublicKey[:])

	n := binary.PutUvarint(temp, delegatedStake.StakeConfirmed)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, delegatedStake.UnstakeHeight)
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
	data, buf, err = helpers.DeserializeBuffer(buf, 33)
	if err != nil {
		return
	}
	copy(delegatedStake.DelegatedPublicKey[:], data)

	delegatedStake.StakeConfirmed, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	delegatedStake.UnstakeHeight, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	var n uint64
	n, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	for i := uint64(0); i < n; i++ {
		var delegatedStakePending = new(DelegatedStakePending)
		buf, err = delegatedStakePending.Deserialize(buf)
		if err != nil {
			return
		}

		delegatedStake.StakesPending = append(delegatedStake.StakesPending, delegatedStakePending)
	}

	out = buf
	return
}
