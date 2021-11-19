package data_storage

import (
	"errors"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
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

	acc, err = dataStorage.CreateAccount(assetId, publicKey)
	if err != nil {
		return nil, nil, err
	}

	return accs, acc, nil
}

func (dataStorage *DataStorage) CreateAccount(assetId, publicKey []byte) (*account.Account, error) {

	exists, err := dataStorage.Regs.Exists(string(publicKey))
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("Can't create Account as it is not Registered")
	}

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return nil, err
	}

	return accs.CreateNewAccount(publicKey)
}

func (dataStorage *DataStorage) GetOrCreatePlainAccount(publicKey []byte, blockHeight uint64) (*plain_account.PlainAccount, error) {
	plainAcc, err := dataStorage.PlainAccs.GetPlainAccount(publicKey, blockHeight)
	if err != nil {
		return nil, err
	}
	if plainAcc != nil {
		return plainAcc, nil
	}
	return dataStorage.CreatePlainAccount(publicKey)
}

func (dataStorage *DataStorage) CreatePlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {

	exists, err := dataStorage.Regs.Exists(string(publicKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Can't create PlainAccount as Registration already exists")
	}

	return dataStorage.PlainAccs.CreateNewPlainAccount(publicKey)
}

func (dataStorage *DataStorage) CreateRegistration(publicKey []byte) (*registration.Registration, error) {

	exists, err := dataStorage.PlainAccs.Exists(string(publicKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Can't register as a plain Account already exists")
	}

	return dataStorage.Regs.CreateNewRegistration(publicKey)
}

func (dataStorage *DataStorage) SubtractUnclaimed(plainAcc *plain_account.PlainAccount, amount, blockHeight uint64) (err error) {
	if err = plainAcc.AddUnclaimed(false, amount); err != nil {
		return
	}
	if plainAcc.AssetFeeLiquidities.HasAssetFeeLiquidities() && plainAcc.Unclaimed < config_asset_fee.GetRequiredAssetFee(blockHeight) {

		for _, assetFeeLiquidity := range plainAcc.AssetFeeLiquidities.List {
			if err = dataStorage.AstsFeeLiquidityCollection.UpdateLiquidity(plainAcc.PublicKey, 0, 0, assetFeeLiquidity.Asset, asset_fee_liquidity.UPDATE_LIQUIDITY_DELETED); err != nil {
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
