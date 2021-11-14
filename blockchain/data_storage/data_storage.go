package data_storage

import (
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/config/config_asset_fee"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type DataStorage struct {
	hash_map.StoreHashMapRepository
	dbTx                       store_db_interface.StoreDBTransactionInterface
	Regs                       *registrations.Registrations
	PlainAccs                  *plain_accounts.PlainAccounts
	AccsCollection             *accounts.AccountsCollection
	Asts                       *assets.Assets
	AstsFeeLiquidityCollection *assets.AssetsFeeLiquidityCollection
}

func (dataStorage *DataStorage) SubtractUnclaimed(plainAcc *plain_account.PlainAccount, amount, blockHeight uint64) (err error) {
	if err = plainAcc.AddUnclaimed(false, amount); err != nil {
		return
	}
	if plainAcc.AssetFeeLiquidities.HasAssetFeeLiquidities() && plainAcc.Unclaimed < config_asset_fee.GetRequiredAssetFee(blockHeight) {

		for _, assetFeeLiquidity := range plainAcc.AssetFeeLiquidities.List {
			if err = dataStorage.AstsFeeLiquidityCollection.UpdateLiquidity(plainAcc.PublicKey, 0, 0, assetFeeLiquidity.AssetId, asset_fee_liquidity.UPDATE_LIQUIDITY_DELETED); err != nil {
				return
			}
		}

		plainAcc.AssetFeeLiquidities.Clear()
	}
	return nil
}

func (dataStorage *DataStorage) GetWhoHasAssetTopLiquidity(assetId []byte, blockHeight uint64) (*plain_account.PlainAccount, error) {
	key, err := dataStorage.AstsFeeLiquidityCollection.GetTopLiquidity(assetId)
	if err != nil || key == nil {
		return nil, err
	}

	return dataStorage.PlainAccs.GetPlainAccount(key, blockHeight)
}

func (dataStorage *DataStorage) GetAssetFeeLiquidityTop(assetId []byte, blockHeight uint64) (*asset_fee_liquidity.AssetFeeLiquidity, error) {

	plainAcc, err := dataStorage.GetWhoHasAssetTopLiquidity(assetId, blockHeight)
	if err != nil || plainAcc == nil {
		return nil, err
	}

	return plainAcc.AssetFeeLiquidities.GetLiquidity(assetId), nil
}

func (dataStorage *DataStorage) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	dataStorage.AccsCollection.SetTx(tx)
	dataStorage.AstsFeeLiquidityCollection.SetTx(tx)
	dataStorage.StoreHashMapRepository.SetTx(tx)
}

func NewDataStorage(dbTx store_db_interface.StoreDBTransactionInterface) (out *DataStorage) {

	out = &DataStorage{
		hash_map.StoreHashMapRepository{},
		dbTx,
		registrations.NewRegistrations(dbTx),
		plain_accounts.NewPlainAccounts(dbTx),
		accounts.NewAccountsCollection(dbTx),
		assets.NewAssets(dbTx),
		assets.NewAssetsFeeLiquidityCollection(dbTx),
	}

	out.GetList = func(computeChangesSize bool) (list []*hash_map.HashMap) {

		list = []*hash_map.HashMap{
			out.Regs.HashMap,
			out.PlainAccs.HashMap,
			out.Asts.HashMap,
		}
		list = append(list, out.AccsCollection.GetAllHashmaps()...)

		if !computeChangesSize {
			list = append(list, out.AstsFeeLiquidityCollection.GetAllHashmaps()...)
		}

		return
	}

	return
}
