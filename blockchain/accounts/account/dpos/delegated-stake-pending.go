package dpos

import (
	"pandora-pay/helpers"
)

type TransactionSimpleScriptType bool

const (
	DelegatedStakePendingStake   TransactionSimpleScriptType = true
	DelegatedStakePendingUnstake TransactionSimpleScriptType = false
)

type DelegatedStakePending struct {
	helpers.SerializableInterface

	PendingAmount    uint64                      //pending stake
	ActivationHeight uint64                      //height when the stake pending was last updated
	PendingType      TransactionSimpleScriptType //true stake pending || false unstake pending
}

func (delegatedStakePending *DelegatedStakePending) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(delegatedStakePending.PendingAmount)
	writer.WriteUvarint(delegatedStakePending.ActivationHeight)
	writer.WriteBool(bool(delegatedStakePending.PendingType))
}

func (delegatedStakePending *DelegatedStakePending) Deserialize(reader *helpers.BufferReader) (err error) {

	if delegatedStakePending.PendingAmount, err = reader.ReadUvarint(); err != nil {
		return
	}
	if delegatedStakePending.ActivationHeight, err = reader.ReadUvarint(); err != nil {
		return
	}

	var read bool
	if read, err = reader.ReadBool(); err != nil {
		return
	}
	delegatedStakePending.PendingType = TransactionSimpleScriptType(read)

	return
}
