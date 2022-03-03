package delegated_pending_stakes

import (
	"errors"
	"pandora-pay/blockchain/data_storage/accounts/account/account_balance_homomorphic"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type DelegatedPendingStakeType bool

type DelegatedPendingStake struct {
	helpers.SerializableInterface `json:"-"  msgpack:"-"`
	PublicKey                     []byte                                          `json:"publicKey" msgpack:"publicKey"`
	PendingAmount                 *account_balance_homomorphic.BalanceHomomorphic `json:"balance" msgpack:"balance"`
}

func (d *DelegatedPendingStake) Validate() error {
	if len(d.PublicKey) != cryptography.PublicKeySize {
		return errors.New("DelegatedPendingStake PublicKey size is invalid")
	}
	return nil
}

func (d *DelegatedPendingStake) Serialize(w *helpers.BufferWriter) {
	w.Write(d.PublicKey)
	d.PendingAmount.Serialize(w)
}

func (d *DelegatedPendingStake) Deserialize(r *helpers.BufferReader) (err error) {
	if d.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if err = d.PendingAmount.Deserialize(r); err != nil {
		return
	}
	return
}
