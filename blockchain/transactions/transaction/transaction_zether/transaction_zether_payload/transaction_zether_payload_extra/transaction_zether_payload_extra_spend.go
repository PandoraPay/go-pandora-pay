package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraSpend struct {
	TransactionZetherPayloadExtraInterface
	SenderSpendPublicKey *bn256.G1
	SenderSpendSignature []byte
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) ComputeAllKeys(out map[string]bool) {
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) VerifyExtraSignature(hashForSignature []byte, payloadStatement *crypto.Statement) bool {
	return crypto.VerifySignaturePoint(hashForSignature, payloadExtra.SenderSpendSignature, payloadExtra.SenderSpendPublicKey)
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, payloadParity bool) error {
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("payloadAsset must be NATIVE_ASSET_FULL")
	}
	if len(payloadExtra.SenderSpendSignature) != cryptography.SignatureSize {
		return errors.New("Invalid Signature size")
	}
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.SenderSpendPublicKey.EncodeCompressed())
	if inclSignature {
		w.Write(payloadExtra.SenderSpendSignature)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.SenderSpendPublicKey, err = r.ReadBN256G1(); err != nil {
		return
	}
	if payloadExtra.SenderSpendSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}

func (payloadExtra *TransactionZetherPayloadExtraSpend) UpdateStatement(payloadStatement *crypto.Statement) error {
	return nil
}
