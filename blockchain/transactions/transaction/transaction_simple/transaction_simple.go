package transaction_simple

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	transaction_base_interface.TransactionBaseInterface
	Extra       transaction_simple_extra.TransactionSimpleExtraInterface
	TxScript    ScriptType
	DataVersion transaction_data.TransactionDataVersion
	Data        []byte
	Nonce       uint64
	Vin         []*transaction_simple_parts.TransactionSimpleInput
	Vout        []*transaction_simple_parts.TransactionSimpleOutput
	Bloom       *TransactionSimpleBloom
}

func (tx *TransactionSimple) IncludeTransaction(blockHeight uint64, txHash []byte, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount
	var acc *account.Account
	var accs *accounts.Accounts

	for i, vin := range tx.Vin {

		if i == 0 {
			if plainAcc, err = dataStorage.GetOrCreatePlainAccount(tx.Bloom.VinPublicKeyHashes[i]); err != nil {
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

			if err = dataStorage.PlainAccs.Update(string(tx.Bloom.VinPublicKeyHashes[i]), plainAcc); err != nil {
				return
			}
		}

		if accs, err = dataStorage.AccsCollection.GetMap(vin.Asset); err != nil {
			return
		}
		if acc, err = accs.GetAccount(tx.Bloom.VinPublicKeyHashes[i]); err != nil {
			return
		}
		if err = acc.AddBalance(false, vin.Amount); err != nil {
			return
		}
		if err = accs.Update(string(tx.Bloom.VinPublicKeyHashes[i]), acc); err != nil {
			return
		}
	}

	for _, vout := range tx.Vout {
		if accs, acc, err = dataStorage.GetOrCreateAccount(vout.Asset, vout.PublicKeyHash); err != nil {
			return
		}
		if err = acc.AddBalance(true, vout.Amount); err != nil {
			return
		}
		if err = accs.Update(string(vout.PublicKeyHash), acc); err != nil {
			return
		}
	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	case SCRIPT_UNSTAKE:
		if err = tx.Extra.IncludeTransactionExtra(blockHeight, tx.Bloom.VinPublicKeyHashes, tx.Vin, tx.Vout, dataStorage); err != nil {
			return
		}
	}

	return nil
}

func (tx *TransactionSimple) ComputeFee() (uint64, error) {
	if err := tx.Bloom.verifyIfBloomed(); err != nil {
		return 0, err
	}
	return tx.Bloom.TransferMap[string(config_coins.NATIVE_ASSET_FULL)], nil
}

func (tx *TransactionSimple) ComputeAllKeys(out map[string]bool) {
	for i := range tx.Vin {
		out[string(tx.Bloom.VinPublicKeyHashes[i])] = true
	}
	for _, vout := range tx.Vout {
		out[string(vout.PublicKeyHash)] = true
	}
	return
}

func (tx *TransactionSimple) VerifySignatureManually(hashForSignature []byte) bool {
	for _, vin := range tx.Vin {
		if !cryptography.VerifySignature(vin.PublicKey, hashForSignature, vin.Signature) {
			return false
		}
	}
	return true
}

func (tx *TransactionSimple) Validate() (err error) {

	if len(tx.Vin) == 0 || len(tx.Vin) > 255 {
		return errors.New("Invalid Vin length")
	}
	if len(tx.Vout) > 255 {
		return errors.New("Invalid Vout length")
	}

	for _, vin := range tx.Vin {
		if err = vin.Validate(); err != nil {
			return
		}
	}

	for _, vout := range tx.Vout {
		if err = vout.Validate(); err != nil {
			return
		}
	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	case SCRIPT_UNSTAKE:
		if tx.Extra == nil {
			return errors.New("extra is not assigned")
		}
		if err = tx.Extra.Validate(tx.Vin, tx.Vout); err != nil {
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
		w.WriteVariableBytes(tx.Data)
	}

	w.WriteUvarint(tx.Nonce)

	w.WriteByte(byte(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(w, inclSignature)
	}

	w.WriteByte(byte(len(tx.Vout)))
	for _, vout := range tx.Vout {
		vout.Serialize(w)
	}

	if tx.Extra != nil {
		tx.Extra.Serialize(w, tx.Vin, tx.Vout, inclSignature)
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
	case SCRIPT_UNSTAKE:
		tx.Extra = &transaction_simple_extra.TransactionSimpleExtraUnstake{}
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

	var c byte
	if c, err = r.ReadByte(); err != nil {
		return
	}

	tx.Vin = make([]*transaction_simple_parts.TransactionSimpleInput, c)
	for i := range tx.Vin {
		tx.Vin[i] = &transaction_simple_parts.TransactionSimpleInput{}
		if err = tx.Vin[i].Deserialize(r); err != nil {
			return
		}
	}

	if c, err = r.ReadByte(); err != nil {
		return
	}
	tx.Vout = make([]*transaction_simple_parts.TransactionSimpleOutput, c)
	for i := range tx.Vout {
		tx.Vout[i] = &transaction_simple_parts.TransactionSimpleOutput{}
		if err = tx.Vout[i].Deserialize(r); err != nil {
			return
		}
	}

	if tx.Extra != nil {
		return tx.Extra.Deserialize(r, tx.Vin, tx.Vout)
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
