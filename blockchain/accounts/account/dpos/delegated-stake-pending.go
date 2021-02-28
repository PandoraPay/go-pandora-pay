package dpos

import (
	"pandora-pay/helpers"
)

type DelegatedStakePending struct {

	//pending stake
	StakePending uint64

	//height when the stake pending was last updated
	StakePendingHeight uint64
}

func (delegatedStakePending *DelegatedStakePending) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUint64(delegatedStakePending.StakePending)
	writer.WriteUint64(delegatedStakePending.StakePendingHeight)

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
