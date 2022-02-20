package dpos

import (
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/helpers"
)

type DelegatedStakePendingType bool

type DelegatedStakePending struct {
	helpers.SerializableInterface `json:"-"  msgpack:"-"`
	PendingAmount                 *account.BalanceHomomorphic `json:"balance" msgpack:"balance"`
	ActivationHeight              uint64                      `json:"activationHeight"  msgpack:"activationHeight"` //height when the stake pending was last updated
}

func (delegatedStakePending *DelegatedStakePending) Serialize(w *helpers.BufferWriter) {
	delegatedStakePending.PendingAmount.Serialize(w)
	w.WriteUvarint(delegatedStakePending.ActivationHeight)
}

func (delegatedStakePending *DelegatedStakePending) Deserialize(r *helpers.BufferReader) (err error) {
	if err = delegatedStakePending.PendingAmount.Deserialize(r); err != nil {
		return
	}
	if delegatedStakePending.ActivationHeight, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
