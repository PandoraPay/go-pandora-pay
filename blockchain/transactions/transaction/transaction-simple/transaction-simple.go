package transaction_simple

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/gui"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	transaction_base_interface.TransactionBaseInterface
	TxScript TransactionSimpleScriptType
	Nonce    uint64
	Vin      []*TransactionSimpleInput
	Vout     []*TransactionSimpleOutput
	Extra    transaction_simple_extra.TransactionSimpleExtraInterface

	Bloom *TransactionSimpleBloom `json:"-"`
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	for i, vin := range tx.Vin {

		var acc *account.Account
		if acc, err = accs.GetAccountEvenEmpty(vin.Bloom.PublicKeyHash, blockHeight); err != nil {
			return
		}

		if i == 0 {
			if acc.Nonce != tx.Nonce {
				return errors.New("Account nonce doesn't match")
			}
			gui.Log("acc.Nonce", acc.Nonce, tx.Nonce)
			if err = acc.IncrementNonce(true); err != nil {
				return
			}

			switch tx.TxScript {
			case TxSimpleScriptDelegate, TxSimpleScriptUnstake:
				if err = tx.Extra.IncludeTransactionVin0(blockHeight, acc); err != nil {
					return
				}
			}
		}

		if err = acc.AddBalance(false, vin.Amount, vin.Token); err != nil {
			return
		}
		accs.UpdateAccount(vin.Bloom.PublicKeyHash, acc)
	}

	for _, vout := range tx.Vout {

		var acc *account.Account
		if acc, err = accs.GetAccountEvenEmpty(vout.PublicKeyHash, blockHeight); err != nil {
			return
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

	if tx.Extra != nil {
		if err = tx.Extra.Validate(); err != nil {
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

func (tx *TransactionSimple) SerializeAdvanced(writer *helpers.BufferWriter, inclSignature bool) {

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

	if tx.Extra != nil {
		tx.Extra.Serialize(writer)
	}
}

func (tx *TransactionSimple) Serialize(writer *helpers.BufferWriter) {
	tx.SerializeAdvanced(writer, true)
}

func (tx *TransactionSimple) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	tx.Serialize(writer)
	return writer.Bytes()
}

func (tx *TransactionSimple) Deserialize(reader *helpers.BufferReader) (err error) {

	var n uint64

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.TxScript = TransactionSimpleScriptType(n)
	switch tx.TxScript {
	case TxSimpleScriptNormal:
		//nothing
	case TxSimpleScriptUnstake, TxSimpleScriptWithdraw:
		tx.Extra = &transaction_simple_extra.TransactionSimpleUnstake{}
	case TxSimpleScriptDelegate:
		tx.Extra = &transaction_simple_extra.TransactionSimpleDelegate{}
	default:
		return errors.New("Invalid TxType")
	}

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

	if tx.Extra != nil {
		return tx.Extra.Deserialize(reader)
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
