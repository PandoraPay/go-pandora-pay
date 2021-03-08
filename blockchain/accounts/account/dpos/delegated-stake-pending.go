package dpos

import (
	"pandora-pay/helpers"
)

type DelegatedStakePending struct {

	//pending stake
	PendingAmount uint64

	//height when the stake pending was last updated
	ActivationHeight uint64
}

func (delegatedStakePending *DelegatedStakePending) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUvarint(delegatedStakePending.PendingAmount)
	writer.WriteUvarint(delegatedStakePending.ActivationHeight)

}

func (delegatedStakePending *DelegatedStakePending) Deserialize(reader *helpers.BufferReader) {

	delegatedStakePending.PendingAmount = reader.ReadUvarint()
	delegatedStakePending.ActivationHeight = reader.ReadUvarint()

}
