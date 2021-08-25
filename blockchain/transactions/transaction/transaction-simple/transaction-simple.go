package transaction_simple

import (
	"bytes"
	"errors"
	"fmt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	"pandora-pay/config"
	"pandora-pay/cryptography/cryptolib"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	transaction_base_interface.TransactionBaseInterface
	transaction_simple_extra.TransactionSimpleExtraInterface
	TxScript ScriptType
	Nonce    uint64
	Token    helpers.HexBytes //20
	Vin      []*transaction_simple_parts.TransactionSimpleInput
	Vout     []*transaction_simple_parts.TransactionSimpleOutput
	Bloom    *TransactionSimpleBloom
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	for i, vin := range tx.Vin {

		var acc *account.Account
		if acc, err = accs.GetAccountEvenEmpty(vin.PublicKey, blockHeight); err != nil {
			return
		}

		if i == 0 {
			if acc.Nonce != tx.Nonce {
				return fmt.Errorf("Account nonce doesn't match %d %d", acc.Nonce, tx.Nonce)
			}
			if err = acc.IncrementNonce(true); err != nil {
				return
			}

			switch tx.TxScript {
			case SCRIPT_DELEGATE, SCRIPT_UNSTAKE:
				if err = tx.TransactionSimpleExtraInterface.IncludeTransactionVin0(blockHeight, acc); err != nil {
					return
				}
			}
		}

		if err = acc.AddBalance(false, vin.Amount, tx.Token); err != nil {
			return
		}
		if err = accs.UpdateAccount(vin.PublicKey, acc); err != nil {
			return
		}
	}

	for _, vout := range tx.Vout {

		var acc *account.Account
		if acc, err = accs.GetAccountEvenEmpty(vout.PublicKey, blockHeight); err != nil {
			return
		}

		if err = acc.AddBalance(true, vout.Amount, tx.Token); err != nil {
			return
		}
		if err = accs.UpdateAccount(vout.PublicKey, acc); err != nil {
			return
		}
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
	case SCRIPT_UNSTAKE:
		return helpers.SafeMapUint64Add(out, config.NATIVE_TOKEN_STRING, tx.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
	}
	return
}

func (tx *TransactionSimple) ComputeAllKeys(out map[string]bool) {
	for _, vin := range tx.Vin {
		out[string(vin.PublicKey)] = true
	}
	for _, vout := range tx.Vout {
		out[string(vout.PublicKey)] = true
	}
	return
}

func (tx *TransactionSimple) ComputeVin(out map[string]uint64) (err error) {
	for _, vin := range tx.Vin {
		if err = helpers.SafeMapUint64Add(out, string(tx.Token), vin.Amount); err != nil {
			return
		}
	}
	return
}

func (tx *TransactionSimple) ComputeVout(out map[string]uint64) (err error) {
	for _, vout := range tx.Vout {
		tokenStr := string(tx.Token)
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
		if cryptolib.VerifySignature(hashForSignature, vin.Signature, vin.PublicKey) == false {
			return false
		}
	}
	return true
}

func (tx *TransactionSimple) Validate() (err error) {

	for _, vin := range tx.Vin {
		if bytes.Equal(vin.PublicKey, config.BURN_PUBLIC_KEY) {
			return errors.New("Input includes BURN ADDR")
		}
	}

	switch tx.TxScript {
	case SCRIPT_NORMAL:
		if len(tx.Vin) == 0 || len(tx.Vin) > 255 {
			return errors.New("Invalid vin")
		}
		if len(tx.Vout) == 0 || len(tx.Vout) > 255 {
			return errors.New("Invalid vout")
		}
	case SCRIPT_DELEGATE, SCRIPT_UNSTAKE:
		if len(tx.Vin) != 1 {
			return errors.New("Invalid vin")
		}
		if len(tx.Vout) != 0 {
			return errors.New("Invalid vout")
		}
	default:
		return errors.New("Invalid TxScript")
	}

	if tx.TransactionSimpleExtraInterface != nil {
		if err = tx.TransactionSimpleExtraInterface.Validate(); err != nil {
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
	writer.WriteToken(tx.Token)

	writer.WriteUvarint(uint64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(writer, inclSignature)
	}

	writer.WriteUvarint(uint64(len(tx.Vout)))
	for _, vout := range tx.Vout {
		vout.Serialize(writer)
	}

	if tx.TransactionSimpleExtraInterface != nil {
		tx.TransactionSimpleExtraInterface.Serialize(writer)
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

	scriptType := ScriptType(n)
	if scriptType >= SCRIPT_END {
		return errors.New("INVALID SCRIPT TYPE")
	}

	tx.TxScript = scriptType
	switch tx.TxScript {
	case SCRIPT_NORMAL:
		//nothing
	case SCRIPT_UNSTAKE:
		tx.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleUnstake{}
	case SCRIPT_DELEGATE:
		tx.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleDelegate{}
	default:
		return errors.New("Invalid TxType")
	}

	if tx.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}
	if tx.Token, err = reader.ReadToken(); err != nil {
		return
	}

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.Vin = make([]*transaction_simple_parts.TransactionSimpleInput, n)
	for i := 0; i < int(n); i++ {
		tx.Vin[i] = &transaction_simple_parts.TransactionSimpleInput{}
		if err = tx.Vin[i].Deserialize(reader); err != nil {
			return
		}
	}

	//vout only TransactionTypeSimple
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.Vout = make([]*transaction_simple_parts.TransactionSimpleOutput, n)
	for i := 0; i < int(n); i++ {
		tx.Vout[i] = &transaction_simple_parts.TransactionSimpleOutput{}
		if err = tx.Vout[i].Deserialize(reader); err != nil {
			return
		}
	}

	if tx.TransactionSimpleExtraInterface != nil {
		return tx.TransactionSimpleExtraInterface.Deserialize(reader)
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
