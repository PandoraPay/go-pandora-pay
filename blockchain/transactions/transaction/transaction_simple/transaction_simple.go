package transaction_simple

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/config"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers/advanced_buffers"
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

	if tx.HasVin() {
		if plainAcc, err = dataStorage.PlainAccs.Get(string(tx.Vin.PublicKey)); err != nil {
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
	}

	if tx.Extra != nil {
		if err = tx.Extra.IncludeTransactionVin0(blockHeight, plainAcc, dataStorage); err != nil {
			return
		}
	}

	if tx.HasVin() {
		return dataStorage.PlainAccs.Update(string(tx.Vin.PublicKey), plainAcc)
	}

	return nil
}

func (tx *TransactionSimple) ComputeFee() (uint64, error) {
	return tx.Fee, nil
}

func (tx *TransactionSimple) ComputeAllKeys(out map[string]bool) {

	if tx.HasVin() {
		out[string(tx.Vin.PublicKey)] = true
	}

	return
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {
	if tx.HasVin() {
		if !crypto.VerifySignature(hashForSignature, tx.Vin.Signature, tx.Vin.PublicKey) {
			return false
		}
	}
	if tx.TxScript == SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT {
		extra := tx.Extra.(*transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment)
		if !extra.VerifySignature() {
			return false
		}
	}

	return true
}

func (tx *TransactionSimple) Validate() (err error) {

	if tx.HasVin() {
		if err = tx.Vin.Validate(); err != nil {
			return
		}
	}

	switch tx.TxScript {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY, SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT, TRANSACTION_SIMPLE_NOTHING:
		if tx.Extra == nil {
			return errors.New("extra is not assigned")
		}
		if err = tx.Extra.Validate(tx.Fee); err != nil {
			return
		}
	default:
		return errors.New("Invalid Simple TxScript")
	}

	return
}

func (tx *TransactionSimple) SerializeAdvanced(w *advanced_buffers.BufferWriter, inclSignature bool) {

	w.WriteUvarint(uint64(tx.TxScript))

	w.WriteByte(byte(tx.DataVersion))
	if tx.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT || tx.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
		w.WriteVariableBytes(tx.Data)
	}

	if tx.HasVin() {
		w.WriteUvarint(tx.Nonce)
		w.WriteUvarint(tx.Fee)
		tx.Vin.Serialize(w, inclSignature)
	}

	if tx.Extra != nil {
		tx.Extra.Serialize(w, inclSignature)
	}
}

func (tx *TransactionSimple) Serialize(w *advanced_buffers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *TransactionSimple) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	tx.TxScript = ScriptType(n)
	switch tx.TxScript {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		tx.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{}
	case SCRIPT_RESOLUTION_CONDITIONAL_PAYMENT:
		tx.Extra = &transaction_simple_extra.TransactionSimpleExtraResolutionConditionalPayment{}
	case TRANSACTION_SIMPLE_NOTHING:
		TX.EXTRA = &transaction_simple_extra.TransactionSimpleNothingt{}
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

	if tx.HasVin() {
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
	}

	if tx.Extra != nil {
		return tx.Extra.Deserialize(r)
	}

	return
}

func (tx *TransactionSimple) HasVin() bool {
	switch tx.TxScript {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		return true
	default:
		return false
	}
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
