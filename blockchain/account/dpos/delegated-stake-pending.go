package dpos

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type DelegatedStakePending struct {

	//pending stake
	StakePending uint64

	//height when the stake pending was last updated
	StakePendingHeight uint64
}

func (delegatedStakePending *DelegatedStakePending) Serialize(serialized *bytes.Buffer, buf []byte) {

	n := binary.PutUvarint(buf, delegatedStakePending.StakePending)
	serialized.Write(buf[:n])

	n = binary.PutUvarint(buf, delegatedStakePending.StakePendingHeight)
	serialized.Write(buf[:n])

}

func (delegatedStakePending *DelegatedStakePending) Deserialize(buf []byte) (out []byte, err error) {

	delegatedStakePending.StakePending, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	delegatedStakePending.StakePendingHeight, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	out = buf
	return
}
