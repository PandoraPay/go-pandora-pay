package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraAssetSupplyIncrease struct {
	TransactionZetherPayloadExtraInterface
	AssetId              []byte
	ReceiverPublicKey    []byte //must be registered before
	Value                uint64
	AssetSupplyPublicKey []byte //TODO: it can be bloomed
	AssetSignature       []byte
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	ast, err := dataStorage.Asts.GetAsset(payloadExtra.AssetId)
	if err != nil {
		return
	}

	if ast == nil {
		return errors.New("Asset was not found")
	}

	if !bytes.Equal(payloadExtra.AssetSupplyPublicKey, ast.SupplyPublicKey) {
		return errors.New("Asset SupplyPublicKey is not matching")
	}

	accs, acc, err := dataStorage.GetOrCreateAccount(payloadExtra.AssetId, payloadExtra.ReceiverPublicKey, blockHeight, true)
	if err != nil {
		return
	}

	if err = ast.AddSupply(true, payloadExtra.Value); err != nil {
		return
	}
	acc.Balance.AddBalanceUint(payloadExtra.Value)

	if err = accs.Update(string(payloadExtra.ReceiverPublicKey), acc); err != nil {
		return
	}
	if err = dataStorage.Asts.Update(string(payloadExtra.AssetId), ast); err != nil {
		return
	}
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) ComputeAllKeys(out map[string]bool) {
	out[string(payloadExtra.ReceiverPublicKey)] = true
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) VerifyExtraSignature(hashForSignature []byte) bool {
	return crypto.VerifySignature(hashForSignature, payloadExtra.AssetSignature, payloadExtra.AssetSupplyPublicKey)
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, payloadParity bool) error {
	if payloadExtra.Value == 0 {
		return errors.New("Asset Supply must be greater than zero")
	}
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("payloadAsset must be NATIVE_ASSET_FULL")
	}
	if len(payloadExtra.ReceiverPublicKey) != cryptography.PublicKeySize || len(payloadExtra.AssetSupplyPublicKey) != cryptography.PublicKeySize {
		return errors.New("Invalid Public Keys")
	}
	if len(payloadExtra.AssetSignature) != cryptography.SignatureSize {
		return errors.New("Invalid Signature")
	}
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.AssetId)
	w.Write(payloadExtra.ReceiverPublicKey)
	w.WriteUvarint(payloadExtra.Value)
	w.Write(payloadExtra.AssetSupplyPublicKey)
	if inclSignature {
		w.Write(payloadExtra.AssetSignature)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.AssetId, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if payloadExtra.ReceiverPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if payloadExtra.Value, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.AssetSupplyPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if payloadExtra.AssetSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) UpdateStatement(payloadStatement *crypto.Statement) error {
	return nil
}
