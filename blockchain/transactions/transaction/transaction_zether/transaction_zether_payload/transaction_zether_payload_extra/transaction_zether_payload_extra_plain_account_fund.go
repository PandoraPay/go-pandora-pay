package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers/advanced_buffers"
)

type TransactionZetherPayloadExtraPlainAccountFund struct {
	TransactionZetherPayloadExtraInterface
	PlainAccountPublicKey []byte
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	var plainAcc *plain_account.PlainAccount
	if plainAcc, err = dataStorage.GetOrCreatePlainAccount(payloadExtra.PlainAccountPublicKey, true); err != nil {
		return err
	}
	if err = plainAcc.AddUnclaimed(true, payloadBurnValue); err != nil {
		return
	}
	return dataStorage.PlainAccs.Update(string(payloadExtra.PlainAccountPublicKey), plainAcc)
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) ComputeAllKeys(out map[string]bool) {
	out[string(payloadExtra.PlainAccountPublicKey)] = true
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) VerifyExtraSignature(hashForSignature []byte, payloadStatement *crypto.Statement) bool {
	return false
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, payloadParity bool) error {
	if len(payloadExtra.PlainAccountPublicKey) != cryptography.PublicKeySize {
		return errors.New("PlainAccountPublicKey size is invalid")
	}
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("PayloadAsset must be NATIVE_ASSET_FULL")
	}
	if payloadBurnValue == 0 {
		return errors.New("Payload Burn value must be greater than zero")
	}
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) Serialize(w *advanced_buffers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.PlainAccountPublicKey)
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if payloadExtra.PlainAccountPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return errors.New("PlainAccountPublicKey was not found")
	}
	return
}

func (payloadExtra *TransactionZetherPayloadExtraPlainAccountFund) UpdateStatement(payloadStatement *crypto.Statement) error {
	return nil
}
