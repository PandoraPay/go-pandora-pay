package dpos

import (
	"pandora-pay/helpers"
)

type DelegatedStakePendingType bool

const (
	DelegatedStakePendingStake   DelegatedStakePendingType = true
	DelegatedStakePendingUnstake DelegatedStakePendingType = false
)

type DelegatedStakePending struct {
	helpers.SerializableInterface `json:"-"`
	PendingAmount                 uint64                    `json:"pendingAmount"`    //pending stake
	ActivationHeight              uint64                    `json:"activationHeight"` //height when the stake pending was last updated
	PendingType                   DelegatedStakePendingType `json:"pendingType"`      //true stake pending || false unstake pending
}

func (delegatedStakePending *DelegatedStakePending) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(delegatedStakePending.PendingAmount)
	w.WriteUvarint(delegatedStakePending.ActivationHeight)
	w.WriteBool(bool(delegatedStakePending.PendingType))
}

func (delegatedStakePending *DelegatedStakePending) Deserialize(r *helpers.BufferReader) (err error) {

	if delegatedStakePending.PendingAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if delegatedStakePending.ActivationHeight, err = r.ReadUvarint(); err != nil {
		return
	}

	var read bool
	if read, err = r.ReadBool(); err != nil {
		return
	}
	delegatedStakePending.PendingType = DelegatedStakePendingType(read)

	return
}
