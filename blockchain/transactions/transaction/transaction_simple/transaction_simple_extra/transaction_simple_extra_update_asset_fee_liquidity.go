package transaction_simple_extra

import (
	"bytes"
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_asset_fee"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraUpdateAssetFeeLiquidity struct {
	TransactionSimpleExtraInterface
	AssetId        []byte
	ConversionRate uint64
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if plainAcc.Unclaimed < config_asset_fee.GetRequiredAssetFee(blockHeight) {
		return fmt.Errorf("Unclaimed must be greater than %d", config_asset_fee.GetRequiredAssetFee(blockHeight))
	}

	if err = plainAcc.UpdateAssetFeeLiquidity(txExtra.AssetId, txExtra.ConversionRate); err != nil {
		return
	}

	return
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) Validate() (err error) {

	if len(txExtra.AssetId) != config_coins.ASSET_LENGTH {
		return errors.New("AssetId length is invalid")
	}

	if bytes.Equal(txExtra.AssetId, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("AssetId NATIVE_ASSET_FULL is not allowed")
	}

	return
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(txExtra.AssetId)
	w.WriteUvarint(txExtra.ConversionRate)
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) Deserialize(r *helpers.BufferReader) (err error) {
	if txExtra.AssetId, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if txExtra.ConversionRate, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
