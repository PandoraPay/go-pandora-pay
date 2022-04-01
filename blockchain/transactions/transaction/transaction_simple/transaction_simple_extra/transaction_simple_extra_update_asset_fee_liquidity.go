package transaction_simple_extra

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/config/config_asset_fee"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraUpdateAssetFeeLiquidity struct {
	TransactionSimpleExtraInterface
	Liquidities     []*asset_fee_liquidity.AssetFeeLiquidity
	CollectorHasNew bool
	Collector       []byte
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if plainAcc.Unclaimed < config_asset_fee.GetRequiredAssetFee(blockHeight) {
		return fmt.Errorf("Unclaimed must be greater than %d", config_asset_fee.GetRequiredAssetFee(blockHeight))
	}

	if txExtra.CollectorHasNew {
		var isReg bool
		if isReg, err = dataStorage.Regs.Exists(string(txExtra.Collector)); err != nil {
			return
		}
		if !isReg {
			return errors.New("Collector must be registered before!")
		}
		plainAcc.AssetFeeLiquidities.Collector = txExtra.Collector
	}

	for _, liquidity := range txExtra.Liquidities {

		var status asset_fee_liquidity.UpdateLiquidityStatus
		if status, err = plainAcc.AssetFeeLiquidities.UpdateLiquidity(liquidity); err != nil {
			return
		}

		if err = dataStorage.AstsFeeLiquidityCollection.UpdateLiquidity(plainAcc.Key, liquidity.Rate, liquidity.LeadingZeros, liquidity.Asset, status); err != nil {
			return
		}

	}

	if len(plainAcc.AssetFeeLiquidities.List) == 0 {
		plainAcc.AssetFeeLiquidities.Collector = nil
		plainAcc.AssetFeeLiquidities.Version = asset_fee_liquidity.NONE
	} else {
		plainAcc.AssetFeeLiquidities.Version = asset_fee_liquidity.SIMPLE
	}

	return
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) Validate() (err error) {

	for _, liquidity := range txExtra.Liquidities {
		if err = liquidity.Validate(); err != nil {
			return
		}
	}

	return
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteBool(txExtra.CollectorHasNew)
	w.Write(txExtra.Collector)

	w.WriteByte(byte(len(txExtra.Liquidities)))
	for _, liquidity := range txExtra.Liquidities {
		liquidity.Serialize(w)
	}
}

func (txExtra *TransactionSimpleExtraUpdateAssetFeeLiquidity) Deserialize(r *helpers.BufferReader) (err error) {

	if txExtra.CollectorHasNew, err = r.ReadBool(); err != nil {
		return
	}

	if txExtra.CollectorHasNew {
		if txExtra.Collector, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
	}

	var count byte
	if count, err = r.ReadByte(); err != nil {
		return
	}

	txExtra.Liquidities = make([]*asset_fee_liquidity.AssetFeeLiquidity, count)
	for i := range txExtra.Liquidities {
		txExtra.Liquidities[i] = &asset_fee_liquidity.AssetFeeLiquidity{}
		if err = txExtra.Liquidities[i].Deserialize(r); err != nil {
			return
		}
	}

	return
}
