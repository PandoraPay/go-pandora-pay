package transaction_simple

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/config"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	transaction_base_interface.TransactionBaseInterface
	Extra       transaction_simple_extra.TransactionSimpleExtraInterface
	TxScript    ScriptType
	DataVersion transaction_data.TransactionDataVersion
	Data        []byte
	Nonce       uint64
	Fee         uint64
	Vin         *transaction_simple_parts.TransactionSimpleInput
	Bloom       *TransactionSimpleBloom
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, txHash []byte, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(tx.Vin.PublicKey); err != nil {
		return
	}
	if plainAcc == nil {
		return errors.New("Plain Account was not found")
	}

	if plainAcc.Nonce != tx.Nonce {
		return fmt.Errorf("Account nonce doesn't match %d %d", plainAcc.Nonce, tx.Nonce)
	}
	if err = plainAcc.IncrementNonce(true); err != nil {
		return
	}

	if err = dataStorage.SubtractUnclaimed(plainAcc, tx.Fee, blockHeight); err != nil {
		return errors.New("Not enought Unclaimed funds to substract Tx.Fee")
	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	}

	return dataStorage.PlainAccs.Update(string(tx.Vin.PublicKey), plainAcc)
}

func (tx *TransactionSimple) ComputeFee() (uint64, error) {
	return tx.Fee, nil
}

func (tx *TransactionSimple) ComputeAllKeys(out map[string]bool) {
	out[string(tx.Vin.PublicKey)] = true
	return
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {
	return ed25519.Verify(tx.Vin.PublicKey, hashForSignature, tx.Vin.Signature)
}

func (tx *TransactionSimple) Validate() (err error) {

	if err = tx.Vin.Validate(); err != nil {
		return
	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	default:
		return errors.New("Invalid Simple TxScript")
	}

	return
}

func (tx *TransactionSimple) SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool) {

	w.WriteUvarint(uint64(tx.TxScript))

	w.WriteByte(byte(tx.DataVersion))
	if tx.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT || tx.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
		w.WriteVariableBytes(tx.Data)
	}

	w.WriteUvarint(tx.Nonce)
	w.WriteUvarint(tx.Fee)

	tx.Vin.Serialize(w, inclSignature)

	if tx.Extra != nil {
		tx.Extra.Serialize(w, inclSignature)
	}
}

func (tx *TransactionSimple) Serialize(w *helpers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *TransactionSimple) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	tx.TxScript = ScriptType(n)
	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	default:
		return errors.New("INVALID SCRIPT TYPE")
	}

	var dataVersion byte
	if dataVersion, err = r.ReadByte(); err != nil {
		return
	}

	tx.DataVersion = transaction_data.TransactionDataVersion(dataVersion)
	switch tx.DataVersion {
	case transaction_data.TX_DATA_NONE:
	case transaction_data.TX_DATA_PLAIN_TEXT, transaction_data.TX_DATA_ENCRYPTED:
		if tx.Data, err = r.ReadVariableBytes(config.TRANSACTIONS_MAX_DATA_LENGTH); err != nil {
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

	if tx.Extra != nil {
		return tx.Extra.Deserialize(r)
	}

	return
}

func (tx *TransactionSimple) VerifyBloomAll() error {
	if tx.Bloom == nil {
		return errors.New("Tx was not bloomed")
	}
	return tx.Bloom.verifyIfBloomed()
}

func (tx *TransactionSimple) GetBloomExtra() any {
	return tx.Bloom
}

func (tx *TransactionSimple) SetBloomExtra(bloom any) {
	if tx.Bloom == nil {
		tx.Bloom = bloom.(*TransactionSimpleBloom)
	}
}
