package pending_stakes

import (
	"pandora-pay/blockchain/data_storage/accounts/account/account_balance_homomorphic"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type PendingStakes struct {
	hash_map.HashMapElementSerializableInterface `json:"-"  msgpack:"-"`
	Key                                          []byte          `json:"-" msgpack:"-"`
	Height                                       uint64          `json:"height" msgpack:"height"`
	Pending                                      []*PendingStake `json:"list" msgpack:"list"`
}

func (d *PendingStakes) IsDeletable() bool {
	return false
}

func (d *PendingStakes) SetKey(key []byte) {
	d.Key = key
}

func (d *PendingStakes) SetIndex(value uint64) {
}

func (d *PendingStakes) GetIndex() uint64 {
	return 0
}

func (d *PendingStakes) Validate() error {
	for _, pending := range d.Pending {
		if err := pending.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *PendingStakes) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(d.Height)

	w.WriteUvarint(uint64(len(d.Pending)))
	for _, pending := range d.Pending {
		pending.Serialize(w)
	}
}

func (d *PendingStakes) Deserialize(r *helpers.BufferReader) (err error) {
	var n uint64

	if d.Height, err = r.ReadUvarint(); err != nil {
		return
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	d.Pending = make([]*PendingStake, n)
	for i := range d.Pending {
		d.Pending[i] = &PendingStake{PendingAmount: &account_balance_homomorphic.BalanceHomomorphic{nil, nil}}
		if err = d.Pending[i].Deserialize(r); err != nil {
			return
		}
	}

	return
}

func NewPendingStakes(key []byte, index uint64) *PendingStakes {
	return &PendingStakes{
		Key:     key,
		Height:  0,
		Pending: []*PendingStake{},
	}
}
