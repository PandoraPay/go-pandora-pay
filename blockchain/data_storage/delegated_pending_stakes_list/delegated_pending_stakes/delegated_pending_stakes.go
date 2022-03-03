package delegated_pending_stakes

import (
	"pandora-pay/blockchain/data_storage/accounts/account/account_balance_homomorphic"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type DelegatedPendingStakes struct {
	hash_map.HashMapElementSerializableInterface `json:"-"  msgpack:"-"`
	Key                                          []byte                   `json:"-" msgpack:"-"`
	Height                                       uint64                   `json:"height" msgpack:"height"`
	Pending                                      []*DelegatedPendingStake `json:"list" msgpack:"list"`
}

func (d *DelegatedPendingStakes) SetKey(key []byte) {
	d.Key = key
}

func (d *DelegatedPendingStakes) SetIndex(value uint64) {
}

func (d *DelegatedPendingStakes) GetIndex() uint64 {
	return 0
}

func (d *DelegatedPendingStakes) Validate() error {
	for _, pending := range d.Pending {
		if err := pending.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (d *DelegatedPendingStakes) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(d.Height)

	w.WriteUvarint(uint64(len(d.Pending)))
	for _, pending := range d.Pending {
		pending.Serialize(w)
	}
}

func (d *DelegatedPendingStakes) Deserialize(r *helpers.BufferReader) (err error) {
	var n uint64

	if d.Height, err = r.ReadUvarint(); err != nil {
		return
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	d.Pending = make([]*DelegatedPendingStake, n)
	for i := range d.Pending {
		d.Pending[i] = &DelegatedPendingStake{PendingAmount: &account_balance_homomorphic.BalanceHomomorphic{nil, nil}}
		if err = d.Pending[i].Deserialize(r); err != nil {
			return
		}
	}

	return
}

func NewDelegatedPendingStakes(key []byte, index uint64) *DelegatedPendingStakes {
	return &DelegatedPendingStakes{
		Key:     key,
		Height:  0,
		Pending: []*DelegatedPendingStake{},
	}
}
