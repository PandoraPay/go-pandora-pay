package dpos

import (
	"pandora-pay/helpers"
)

type DelegatedStakePending struct {
	PendingAmount    uint64 //pending stake
	ActivationHeight uint64 //height when the stake pending was last updated
	PendingType      bool   //true stake pending || false unstake pending
}

func (delegatedStakePending *DelegatedStakePending) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(delegatedStakePending.PendingAmount)
	writer.WriteUvarint(delegatedStakePending.ActivationHeight)
	writer.WriteBool(delegatedStakePending.PendingType)
}

func (delegatedStakePending *DelegatedStakePending) Deserialize(reader *helpers.BufferReader) {
	delegatedStakePending.PendingAmount = reader.ReadUvarint()
	delegatedStakePending.ActivationHeight = reader.ReadUvarint()
	delegatedStakePending.PendingType = reader.ReadBool()
}
