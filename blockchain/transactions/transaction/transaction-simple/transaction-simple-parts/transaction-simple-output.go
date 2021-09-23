package transaction_simple_parts

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleOutput struct {
	Amount    uint64
	PublicKey []byte
}

func (vout *TransactionSimpleOutput) Validate() error {
	if vout.Amount == 0 {
		return errors.New("Amount must be greater than zero")
	}
	if len(vout.PublicKey) != cryptography.PublicKeySize {
		return errors.New("PublicKey length is invalid")
	}
	return nil
}

func (vout *TransactionSimpleOutput) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(vout.Amount)
	w.Write(vout.PublicKey)
}

func (vout *TransactionSimpleOutput) Deserialize(r *helpers.BufferReader) (err error) {
	if vout.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	if vout.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	return
}
