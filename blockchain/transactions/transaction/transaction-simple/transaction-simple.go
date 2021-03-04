package transaction_simple

import (
	"errors"
	"math"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction_simple_unstake"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	Nonce uint64
	Vin   []*TransactionSimpleInput
	Vout  []*TransactionSimpleOutput
	Extra interface{}
}

func (tx *TransactionSimple) ComputeFees(out map[string]uint64, txType transaction_type.TransactionType) (err error) {
	if err = tx.ComputeVin(out); err != nil {
		return
	}
	if err = tx.ComputeVout(out); err != nil {
		return
	}
	switch txType {
	case transaction_type.TransactionTypeSimpleUnstake:
		extra := tx.Extra.(*transaction_simple_unstake.TransactionSimpleUnstake)
		if math.MaxUint64-out[string(config.NATIVE_TOKEN)] < extra.UnstakeFeeExtra {
			return errors.New("Unstake exceeded MaxUint64")
		}
		out[string(config.NATIVE_TOKEN)] += extra.UnstakeFeeExtra
	}
	return
}

func (tx *TransactionSimple) ComputeVin(out map[string]uint64) error {
	for _, vin := range tx.Vin {
		token := string(vin.Token)
		if math.MaxUint64-out[token] <= vin.Amount {
			return errors.New("Vin exceeded MaxUint64")
		}
		out[token] += vin.Amount
	}
	return nil
}

func (tx *TransactionSimple) ComputeVout(out map[string]uint64) error {
	for _, vout := range tx.Vout {
		token := string(vout.Token)
		if out[token] < vout.Amount {
			return errors.New("Balance exceeded")
		}
		out[token] -= vout.Amount
		if out[token] == 0 {
			delete(out, token)
		}
	}
	return nil
}

func (tx *TransactionSimple) VerifySignature(hash helpers.Hash) bool {
	if len(tx.Vin) == 0 {
		return false
	}

	for _, vin := range tx.Vin {
		if ecdsa.VerifySignature(vin.PublicKey[:], hash[:], vin.Signature[0:64]) == false {
			return false
		}
	}
	return true
}

func (tx *TransactionSimple) Validate(txType transaction_type.TransactionType) (err error) {

	switch txType {
	case transaction_type.TransactionTypeSimple:
		if len(tx.Vin) == 0 || len(tx.Vin) > 255 {
			return errors.New("Invalid vin")
		}
		if len(tx.Vout) == 0 || len(tx.Vout) > 255 {
			return errors.New("Invalid vout")
		}
	case transaction_type.TransactionTypeSimpleUnstake:
		if len(tx.Vin) != 1 {
			return errors.New("Invalid vin")
		}
		if len(tx.Vout) != 0 {
			return errors.New("Invalid vout")
		}
		extra := tx.Extra.(*transaction_simple_unstake.TransactionSimpleUnstake)
		extra.Validate(txType)
	}

	final := make(map[string]uint64)
	if err = tx.ComputeVin(final); err != nil {
		return
	}
	if err = tx.ComputeVout(final); err != nil {
		return
	}

	return
}

func (tx *TransactionSimple) Serialize(writer *helpers.BufferWriter, inclSignature bool, txType transaction_type.TransactionType) {
	writer.WriteUvarint(tx.Nonce)

	writer.WriteUvarint(uint64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(writer, inclSignature)
	}

	writer.WriteUvarint(uint64(len(tx.Vout)))
	for _, vout := range tx.Vout {
		vout.Serialize(writer)
	}

	switch txType {
	case transaction_type.TransactionTypeSimpleUnstake:
		extra := tx.Extra.(*transaction_simple_unstake.TransactionSimpleUnstake)
		extra.Serialize(writer)
	}
}

func (tx *TransactionSimple) Deserialize(reader *helpers.BufferReader, txType transaction_type.TransactionType) (err error) {

	var n uint64

	if tx.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	for i := 0; i < int(n); i++ {
		vin := &TransactionSimpleInput{}
		if err = vin.Deserialize(reader); err != nil {
			return
		}
		tx.Vin = append(tx.Vin, vin)
	}

	//vout only TransactionTypeSimple
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	for i := 0; i < int(n); i++ {
		vout := &TransactionSimpleOutput{}
		if err = vout.Deserialize(reader); err != nil {
			return
		}
		tx.Vout = append(tx.Vout, vout)
	}

	switch txType {
	case transaction_type.TransactionTypeSimpleUnstake:
		extra := &transaction_simple_unstake.TransactionSimpleUnstake{}
		if err = extra.Deserialize(reader); err != nil {
			return err
		}
		tx.Extra = extra
	}

	return
}
