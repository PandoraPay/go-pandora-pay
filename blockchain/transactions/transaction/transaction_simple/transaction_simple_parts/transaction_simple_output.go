package transaction_simple_parts

import (
	"errors"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleOutput struct {
	PublicKeyHash []byte
	Amount        uint64
	Asset         []byte
}

func (vout *TransactionSimpleOutput) Validate() error {
	if vout.Amount == 0 {
		return errors.New("Amount must be greater than zero")
	}
	if len(vout.PublicKeyHash) != cryptography.PublicKeyHashSize {
		return errors.New("PublicKeyHash length is invalid")
	}
	if len(vout.Asset) != config_coins.ASSET_LENGTH {
		return errors.New("Vout.Asset is invalid")
	}
	return nil
}

func (vout *TransactionSimpleOutput) Serialize(w *helpers.BufferWriter) {
	w.Write(vout.PublicKeyHash)
	w.WriteUvarint(vout.Amount)
	w.WriteAsset(vout.Asset)
}

func (vout *TransactionSimpleOutput) Deserialize(r *helpers.BufferReader) (err error) {
	if vout.PublicKeyHash, err = r.ReadBytes(cryptography.PublicKeyHashSize); err != nil {
		return
	}
	if vout.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	if vout.Asset, err = r.ReadAsset(); err != nil {
		return
	}
	return
}
