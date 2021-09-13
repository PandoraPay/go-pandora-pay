package transaction_simple

import (
	"bytes"
	"errors"
	"fmt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	"pandora-pay/config"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	transaction_base_interface.TransactionBaseInterface
	transaction_simple_extra.TransactionSimpleExtraInterface
	TxScript    ScriptType
	DataVersion transaction_data.TransactionDataVersion
	Data        []byte
	Nonce       uint64
	Fee         uint64
	Vin         *transaction_simple_parts.TransactionSimpleInput
	Bloom       *TransactionSimpleBloom
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) (err error) {

	accs, err := accsCollection.GetMap(config.NATIVE_TOKEN_FULL)
	if err != nil {
		return
	}

	var acc *account.Account
	if acc, err = accs.GetAccount(tx.Vin.PublicKey, blockHeight); err != nil {
		return
	}

	if acc == nil {
		return errors.New("Account was not found")
	}

	if acc.NativeExtra.Nonce != tx.Nonce {
		return fmt.Errorf("Account nonce doesn't match %d %d", acc.NativeExtra.Nonce, tx.Nonce)
	}
	if err = acc.NativeExtra.IncrementNonce(true); err != nil {
		return
	}

	if err = acc.NativeExtra.DelegatedStake.AddStakeAvailable(false, tx.Fee); err != nil {
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
	if crypto.VerifySignature(hashForSignature, tx.Vin.Signature, tx.Vin.PublicKey) == false {
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

func (tx *TransactionSimple) SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool) {

	w.WriteUvarint(uint64(tx.TxScript))

	w.WriteByte(byte(tx.DataVersion))
	if tx.DataVersion != transaction_data.TX_DATA_NONE {
		w.WriteUvarint(uint64(len(tx.Data)))
		w.Write(tx.Data)
	}

	w.WriteUvarint(tx.Nonce)
	w.WriteUvarint(tx.Fee)

	tx.Vin.Serialize(w, inclSignature)

	if tx.TransactionSimpleExtraInterface != nil {
		tx.TransactionSimpleExtraInterface.Serialize(w)
	}
}

func (tx *TransactionSimple) Serialize(w *helpers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *TransactionSimple) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	tx.Serialize(w)
	return w.Bytes()
}

func (tx *TransactionSimple) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64

	if n, err = r.ReadUvarint(); err != nil {
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

	var dataVersion byte
	if dataVersion, err = r.ReadByte(); err != nil {
		return
	}

	tx.DataVersion = transaction_data.TransactionDataVersion(dataVersion)
	switch tx.DataVersion {
	case transaction_data.TX_DATA_NONE:
	case transaction_data.TX_DATA_PLAIN_TEXT, transaction_data.TX_DATA_ENCRYPTED:
		if n, err = r.ReadUvarint(); err != nil {
			return
		}
		if n == 0 || n > config.TRANSACTIONS_MAX_DATA_LENGTH {
			return errors.New("Tx.Data length is invalid")
		}
		if tx.Data, err = r.ReadBytes(int(n)); err != nil {
			return
		}
	default:
		return errors.New("Invalid Tx.DataVersion")
	}

	if tx.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}

	if tx.Fee, err = r.ReadUvarint(); err != nil {
		return
	}

	tx.Vin = &transaction_simple_parts.TransactionSimpleInput{}
	if err = tx.Vin.Deserialize(r); err != nil {
		return
	}

	if tx.TransactionSimpleExtraInterface != nil {
		return tx.TransactionSimpleExtraInterface.Deserialize(r)
	}

	return
}

func (tx *TransactionSimple) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
