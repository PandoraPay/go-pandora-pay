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
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	transaction_base_interface.TransactionBaseInterface
	Extra       transaction_simple_extra.TransactionSimpleExtraInterface
	TxScript    ScriptType
	DataVersion transaction_data.TransactionDataVersion
	Data        []byte
	Nonce       uint64
	Fees        uint64
	Vin         *transaction_simple_parts.TransactionSimpleInput
	Bloom       *TransactionSimpleBloom
}

func (tx *TransactionSimple) ComputeExtraSpace() uint64 {
	return 0
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, txHash []byte, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(tx.Vin.PublicKey, blockHeight); err != nil {
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

	switch tx.TxScript {
	case SCRIPT_UPDATE_DELEGATE, SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		if plainAcc.Unclaimed >= tx.Fees {
			err = plainAcc.AddUnclaimed(false, tx.Fees)
		} else {
			err = plainAcc.DelegatedStake.AddStakeAvailable(false, tx.Fees)
		}
	case SCRIPT_UNSTAKE:
		err = plainAcc.DelegatedStake.AddStakeAvailable(false, tx.Fees)
	default:
		err = errors.New("Invalid Simple TxScript")
	}

	if err != nil {
		return
	}

	switch tx.TxScript {
	case SCRIPT_UPDATE_DELEGATE, SCRIPT_UNSTAKE, SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		if err = tx.Extra.IncludeTransactionVin0(blockHeight, plainAcc, dataStorage); err != nil {
			return
		}
	}

	if err = dataStorage.PlainAccs.Update(string(tx.Vin.PublicKey), plainAcc); err != nil {
		return
	}

	return
}

func (tx *TransactionSimple) ComputeFees() (uint64, error) {
	return tx.Fees, nil
}

func (tx *TransactionSimple) ComputeAllKeys(out map[string]bool) {
	out[string(tx.Vin.PublicKey)] = true
	return
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {
	return crypto.VerifySignature(hashForSignature, tx.Vin.Signature, tx.Vin.PublicKey)
}

func (tx *TransactionSimple) Validate() (err error) {

	if err = tx.Vin.Validate(); err != nil {
		return
	}

	switch tx.TxScript {
	case SCRIPT_UPDATE_DELEGATE, SCRIPT_UNSTAKE:
		if tx.Extra == nil {
			return errors.New("extra is not assigned")
		}
		if err = tx.Extra.Validate(); err != nil {
			return
		}
	default:
		return errors.New("Invalid Simple TxScript")
	}

	return
}

func (tx *TransactionSimple) SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool) {

	w.WriteUvarint(uint64(tx.TxScript))

	w.WriteByte(byte(tx.DataVersion))
	if tx.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT || tx.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
		w.WriteUvarint(uint64(len(tx.Data)))
		w.Write(tx.Data)
	}

	w.WriteUvarint(tx.Nonce)
	w.WriteUvarint(tx.Fees)

	tx.Vin.Serialize(w, inclSignature)

	if tx.Extra != nil {
		tx.Extra.Serialize(w, inclSignature)
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

	tx.TxScript = ScriptType(n)
	switch tx.TxScript {
	case SCRIPT_UNSTAKE:
		tx.Extra = &transaction_simple_extra.TransactionSimpleExtraUnstake{}
	case SCRIPT_UPDATE_DELEGATE:
		tx.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateDelegate{}
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		tx.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{}
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

	if tx.Fees, err = r.ReadUvarint(); err != nil {
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

func (tx *TransactionSimple) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
