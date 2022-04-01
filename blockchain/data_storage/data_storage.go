package data_storage

import (
	"errors"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/pending_stakes_list"
	"pandora-pay/blockchain/data_storage/pending_stakes_list/pending_stakes"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/config/config_asset_fee"
	"pandora-pay/config/config_coins"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type DataStorage struct {
	DBTx                       store_db_interface.StoreDBTransactionInterface
	PlainAccs                  *plain_accounts.PlainAccounts
	AccsCollection             *accounts.AccountsCollection
	PendingStakes              *pending_stakes_list.PendingStakesList
	Asts                       *assets.Assets
	AstsFeeLiquidityCollection *assets.AssetsFeeLiquidityCollection
}

func (dataStorage *DataStorage) GetOrCreateAccount(assetId, publicKey []byte) (*accounts.Accounts, *account.Account, error) {

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return nil, nil, err
	}

	acc, err := accs.GetAccount(publicKey)
	if err != nil {
		return nil, nil, err
	}

	if acc != nil {
		return accs, acc, nil
	}

	if acc, err = accs.CreateNewAccount(publicKey); err != nil {
		return nil, nil, err
	}

	return accs, acc, nil
}

func (dataStorage *DataStorage) CreateAccount(assetId, publicKey []byte) (*accounts.Accounts, *account.Account, error) {

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return nil, nil, err
	}

	exists, err := accs.Exists(string(publicKey))
	if err != nil {
		return nil, nil, err
	}

	if exists {
		return nil, nil, errors.New("Account already exists")
	}

	acc, err := accs.CreateNewAccount(publicKey)
	if err != nil {
		return nil, nil, err
	}

	return accs, acc, nil
}

func (dataStorage *DataStorage) GetOrCreatePlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {
	plainAcc, err := dataStorage.PlainAccs.GetPlainAccount(publicKey)
	if err != nil {
		return nil, err
	}
	if plainAcc != nil {
		return plainAcc, nil
	}
	return dataStorage.CreatePlainAccount(publicKey)
}

func (dataStorage *DataStorage) CreatePlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {
	return dataStorage.PlainAccs.CreateNewPlainAccount(publicKey)
}

func (dataStorage *DataStorage) AddStakePendingStake(publicKey []byte, amount uint64, pendingType bool, blockHeight uint64) error {

	pendingStakes, err := dataStorage.PendingStakes.GetPendingStakes(blockHeight)
	if err != nil {
		return err
	}

	if pendingStakes == nil {
		if pendingStakes, err = dataStorage.PendingStakes.CreateNewPendingStakes(blockHeight); err != nil {
			return err
		}
	}

	pendingStakes.Pending = append(pendingStakes.Pending, &pending_stakes.PendingStake{
		PublicKey:     publicKey,
		PendingAmount: amount,
		PendingType:   pendingType,
	})

	return dataStorage.PendingStakes.Update(strconv.FormatUint(blockHeight, 10), pendingStakes)
}

func (dataStorage *DataStorage) ProcessPendingStakes(blockHeight uint64) error {

	accs, err := dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL)
	if err != nil {
		return err
	}

	pendingStakes, err := dataStorage.PendingStakes.GetPendingStakes(blockHeight)
	if err != nil {
		return err
	}

	if pendingStakes == nil {
		return nil
	}

	for _, pending := range pendingStakes.Pending {

		var acc *account.Account
		if acc, err = accs.GetAccount(pending.PublicKey); err != nil {
			return err
		}

		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		panic("todo")

		if err = accs.Update(string(pending.PublicKey), acc); err != nil {
			return err
		}
	}

	dataStorage.PendingStakes.Delete(strconv.FormatUint(blockHeight, 10))
	return nil
}

func (dataStorage *DataStorage) SubtractUnclaimed(plainAcc *plain_account.PlainAccount, amount, blockHeight uint64) (err error) {

	if err = plainAcc.AddUnclaimed(false, amount); err != nil {
		return
	}

	if plainAcc.AssetFeeLiquidities.HasAssetFeeLiquidities() && plainAcc.Unclaimed < config_asset_fee.GetRequiredAssetFee(blockHeight) {

		for _, assetFeeLiquidity := range plainAcc.AssetFeeLiquidities.List {
			if err = dataStorage.AstsFeeLiquidityCollection.UpdateLiquidity(plainAcc.Key, 0, 0, assetFeeLiquidity.Asset, asset_fee_liquidity.UPDATE_LIQUIDITY_DELETED); err != nil {
				return
			}
		}

		plainAcc.AssetFeeLiquidities.Clear()
	}
	return
}

func (dataStorage *DataStorage) GetWhoHasAssetTopLiquidity(assetId []byte) (*plain_account.PlainAccount, error) {
	key, err := dataStorage.AstsFeeLiquidityCollection.GetTopLiquidity(assetId)
	if err != nil || key == nil {
		return nil, err
	}

	return dataStorage.PlainAccs.GetPlainAccount(key)
}

func (dataStorage *DataStorage) GetAssetFeeLiquidityTop(assetId []byte) (*asset_fee_liquidity.AssetFeeLiquidity, error) {

	plainAcc, err := dataStorage.GetWhoHasAssetTopLiquidity(assetId)
	if err != nil || plainAcc == nil {
		return nil, err
	}

	return plainAcc.AssetFeeLiquidities.GetLiquidity(assetId), nil
}

func NewDataStorage(dbTx store_db_interface.StoreDBTransactionInterface) (out *DataStorage) {

	out = &DataStorage{
		dbTx,
		plain_accounts.NewPlainAccounts(dbTx),
		accounts.NewAccountsCollection(dbTx),
		pending_stakes_list.NewPendingStakesList(dbTx),
		assets.NewAssets(dbTx),
		assets.NewAssetsFeeLiquidityCollection(dbTx),
	}

	return
}
