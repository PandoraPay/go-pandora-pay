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

func (delegatedStakePending *DelegatedStakePending) Serialize(serialized *bytes.Buffer, temp []byte) {

	n := binary.PutUvarint(temp, delegatedStakePending.StakePending)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, delegatedStakePending.StakePendingHeight)
	serialized.Write(temp[:n])

}

func (delegatedStakePending *DelegatedStakePending) Deserialize(reader *helpers.BufferReader) (err error) {

	if delegatedStakePending.StakePending, err = reader.ReadUvarint(); err != nil {
		return
	}

	if delegatedStakePending.StakePendingHeight, err = reader.ReadUvarint(); err != nil {
		return
	}

	return
}
