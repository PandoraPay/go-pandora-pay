package transaction_simple_parts

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleOutput struct {
	Amount        uint64
	PublicKeyHash []byte
}

func (vout *TransactionSimpleOutput) Validate() error {
	if vout.Amount == 0 {
		return errors.New("Amount must be greater than zero")
	}
	if len(vout.PublicKeyHash) != cryptography.PublicKeyHashSize {
		return errors.New("PublicKeyHash length is invalid")
	}
	return nil
}

func (vout *TransactionSimpleOutput) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(vout.Amount)
	w.Write(vout.PublicKeyHash)
}

func (vout *TransactionSimpleOutput) Deserialize(r *helpers.BufferReader) (err error) {
	if vout.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	if vout.PublicKeyHash, err = r.ReadBytes(cryptography.PublicKeyHashSize); err != nil {
		return
	}
	return
}
