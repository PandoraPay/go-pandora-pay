package transaction_simple

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	TxScript TransactionSimpleScriptType
	Nonce    uint64
	Vin      []*TransactionSimpleInput
	Vout     []*TransactionSimpleOutput
	Extra    interface{}

	Bloom *TransactionSimpleBloom
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	for i, vin := range tx.Vin {

		acc := accs.GetAccountEvenEmpty(vin.Bloom.PublicKeyHash)
		if err = acc.RefreshDelegatedStake(blockHeight); err != nil {
			return
		}

		if i == 0 {
			if acc.Nonce != tx.Nonce {
				return errors.New("Account nonce doesn't match")
			}
			if err = acc.IncrementNonce(true); err != nil {
				return
			}

			switch tx.TxScript {
			case TxSimpleScriptDelegate:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleDelegate).IncludeTransactionVin0(blockHeight, acc)
			case TxSimpleScriptUnstake:
				tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).IncludeTransactionVin0(blockHeight, acc)
			}
		}

		if err = acc.AddBalance(false, vin.Amount, vin.Token); err != nil {
			return
		}
		accs.UpdateAccount(vin.Bloom.PublicKeyHash, acc)
	}

	for _, vout := range tx.Vout {
		acc := accs.GetAccountEvenEmpty(vout.PublicKeyHash)
		if err = acc.RefreshDelegatedStake(blockHeight); err != nil {
		}

		if err = acc.AddBalance(true, vout.Amount, vout.Token); err != nil {
			return
		}
		accs.UpdateAccount(vout.PublicKeyHash, acc)
	}

	return nil
}

func (tx *TransactionSimple) ComputeFees(out map[string]uint64) (err error) {

	if err = tx.ComputeVin(out); err != nil {
		return
	}
	if err = tx.ComputeVout(out); err != nil {
		return
	}

	switch tx.TxScript {
	case TxSimpleScriptUnstake:
		return helpers.SafeMapUint64Add(out, config.NATIVE_TOKEN_STRING, tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
	}
	return
}

func (tx *TransactionSimple) ComputeVin(out map[string]uint64) (err error) {
	for _, vin := range tx.Vin {
		if err = helpers.SafeMapUint64Add(out, string(vin.Token), vin.Amount); err != nil {
			return
		}
	}
	return
}

func (tx *TransactionSimple) ComputeVout(out map[string]uint64) (err error) {
	for _, vout := range tx.Vout {
		tokenStr := string(vout.Token)
		if err = helpers.SafeMapUint64Sub(out, tokenStr, vout.Amount); err != nil {
			return
		}
		if out[tokenStr] == 0 {
			delete(out, tokenStr)
		}
	}
	return
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {

	if len(tx.Vin) == 0 {
		return false
	}

	for _, vin := range tx.Vin {
		if ecdsa.VerifySignature(vin.Bloom.PublicKey, hashForSignature, vin.Signature[0:64]) == false {
			return false
		}
	}
	return true
}

func (tx *TransactionSimple) Validate() (err error) {

	for _, vin := range tx.Vin {
		if bytes.Equal(vin.Bloom.PublicKeyHash, config.BURN_PUBLIC_KEY_HASH) {
			return errors.New("Input includes BURN ADDR")
		}
	}

	switch tx.TxScript {
	case TxSimpleScriptNormal:
		if len(tx.Vin) == 0 || len(tx.Vin) > 255 {
			return errors.New("Invalid vin")
		}
		if len(tx.Vout) == 0 || len(tx.Vout) > 255 {
			return errors.New("Invalid vout")
		}
	case TxSimpleScriptDelegate, TxSimpleScriptUnstake, TxSimpleScriptWithdraw:
		if len(tx.Vin) != 1 {
			return errors.New("Invalid vin")
		}
		if len(tx.Vout) != 0 {
			return errors.New("Invalid vout")
		}
	default:
		return errors.New("Invalid TxScript")
	}

	switch tx.TxScript {
	case TxSimpleScriptDelegate:
		if err = tx.Extra.(*transaction_simple_extra.TransactionSimpleDelegate).Validate(); err != nil {
			return
		}
	case TxSimpleScriptUnstake:
		if err = tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).Validate(); err != nil {
			return
		}
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

func (tx *TransactionSimple) Serialize(writer *helpers.BufferWriter, inclSignature bool) {

	writer.WriteUvarint(uint64(tx.TxScript))
	writer.WriteUvarint(tx.Nonce)

	writer.WriteUvarint(uint64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(writer, inclSignature)
	}

	writer.WriteUvarint(uint64(len(tx.Vout)))
	for _, vout := range tx.Vout {
		vout.Serialize(writer)
	}

	switch tx.TxScript {
	case TxSimpleScriptDelegate:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleDelegate).Serialize(writer)
	case TxSimpleScriptUnstake:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).Serialize(writer)
	case TxSimpleScriptWithdraw:
		tx.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).Serialize(writer)
	}
}

func (tx *TransactionSimple) Deserialize(reader *helpers.BufferReader) (err error) {

	var n uint64

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.TxScript = TransactionSimpleScriptType(n)
	if tx.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.Vin = make([]*TransactionSimpleInput, n)
	for i := 0; i < int(n); i++ {
		tx.Vin[i] = &TransactionSimpleInput{}
		if err = tx.Vin[i].Deserialize(reader); err != nil {
			return
		}
	}

	//vout only TransactionTypeSimple
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.Vout = make([]*TransactionSimpleOutput, n)
	for i := 0; i < int(n); i++ {
		tx.Vout[i] = &TransactionSimpleOutput{}
		if err = tx.Vout[i].Deserialize(reader); err != nil {
			return
		}
	}

	switch tx.TxScript {
	case TxSimpleScriptDelegate:
		extra := &transaction_simple_extra.TransactionSimpleDelegate{}
		if err = extra.Deserialize(reader); err != nil {
			return
		}
		tx.Extra = extra
	case TxSimpleScriptUnstake:
		extra := &transaction_simple_extra.TransactionSimpleUnstake{}
		if err = extra.Deserialize(reader); err != nil {
			return
		}
		tx.Extra = extra
	}

	return
}

func (tx *TransactionSimple) VerifyBloomAll() (err error) {
	for _, vin := range tx.Vin {
		if err = vin.Bloom.VerifyIfBloomed(); err != nil {
			return
		}
	}
	return tx.Bloom.verifyIfBloomed()
}
