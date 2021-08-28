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
	Fee      uint64
	Vin      *transaction_simple_parts.TransactionSimpleInput
	Bloom    *TransactionSimpleBloom
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	var acc *account.Account
	if acc, err = accs.GetAccount(tx.Vin.PublicKey, blockHeight); err != nil {
		return
	}

	if acc == nil {
		return errors.New("Account was not found")
	}

	if acc.Nonce != tx.Nonce {
		return fmt.Errorf("Account nonce doesn't match %d %d", acc.Nonce, tx.Nonce)
	}
	if err = acc.IncrementNonce(true); err != nil {
		return
	}

	if err = acc.DelegatedStake.AddStakeAvailable(false, tx.Fee); err != nil {
		return
	}

	switch tx.TxScript {
	case SCRIPT_UPDATE_DELEGATE, SCRIPT_UNSTAKE:
		if err = tx.TransactionSimpleExtraInterface.IncludeTransactionVin0(blockHeight, acc); err != nil {
			return
		}
	}

	if err = accs.UpdateAccount(tx.Vin.PublicKey, acc); err != nil {
		return
	}

	return nil
}

func (tx *TransactionSimple) ComputeFees() uint64 {
	return tx.Fee
}

func (tx *TransactionSimple) ComputeAllKeys(out map[string]bool) {
	out[string(tx.Vin.PublicKey)] = true
	return
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {
	if cryptolib.VerifySignature(hashForSignature, tx.Vin.Signature, tx.Vin.PublicKey) == false {
		return false
	}
	return true
}

func (tx *TransactionSimple) Validate() (err error) {

	if bytes.Equal(tx.Vin.PublicKey, config.BURN_PUBLIC_KEY) {
		return errors.New("Input includes BURN ADDR")
	}

	switch tx.TxScript {
	case SCRIPT_UPDATE_DELEGATE, SCRIPT_UNSTAKE:
		if tx.TransactionSimpleExtraInterface == nil {
			return errors.New("extra is not assigned")
		}
		if err = tx.TransactionSimpleExtraInterface.Validate(); err != nil {
			return
		}
	default:
		return errors.New("Invalid TxScript")
	}

	return
}

func (tx *TransactionSimple) SerializeAdvanced(writer *helpers.BufferWriter, inclSignature bool) {

	writer.WriteUvarint(uint64(tx.TxScript))
	writer.WriteUvarint(tx.Nonce)

	tx.Vin.Serialize(writer, inclSignature)

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
	case SCRIPT_UNSTAKE:
		tx.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleUnstake{}
	case SCRIPT_UPDATE_DELEGATE:
		tx.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleUpdateDelegate{}
	default:
		return errors.New("Invalid TxType")
	}

	if tx.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.Vin = &transaction_simple_parts.TransactionSimpleInput{}
	if err = tx.Vin.Deserialize(reader); err != nil {
		return
	}

	if tx.TransactionSimpleExtraInterface != nil {
		return tx.TransactionSimpleExtraInterface.Deserialize(reader)
	}

	return
}

func (tx *TransactionSimple) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
