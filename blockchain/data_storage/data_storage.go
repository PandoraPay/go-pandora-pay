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
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type DataStorage struct {
	DBTx           store_db_interface.StoreDBTransactionInterface
	PlainAccs      *plain_accounts.PlainAccounts
	AccsCollection *accounts.AccountsCollection
	PendingStakes  *pending_stakes_list.PendingStakesList
	Asts           *assets.Assets
}

func (dataStorage *DataStorage) GetOrCreateAccount(assetId, publicKeyHash []byte) (*accounts.Accounts, *account.Account, error) {

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return nil, nil, err
	}

	acc, err := accs.GetAccount(publicKeyHash)
	if err != nil {
		return nil, nil, err
	}

	if acc != nil {
		return accs, acc, nil
	}

	if acc, err = accs.CreateNewAccount(publicKeyHash); err != nil {
		return nil, nil, err
	}

	return accs, acc, nil
}

func (dataStorage *DataStorage) CreateAccount(assetId, publicKeyHash []byte) (*accounts.Accounts, *account.Account, error) {

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return nil, nil, err
	}

	exists, err := accs.Exists(string(publicKeyHash))
	if err != nil {
		return nil, nil, err
	}

	if exists {
		return nil, nil, errors.New("Account already exists")
	}

	acc, err := accs.CreateNewAccount(publicKeyHash)
	if err != nil {
		return nil, nil, err
	}

	return accs, acc, nil
}

func (dataStorage *DataStorage) GetOrCreatePlainAccount(publicKeyHash []byte) (*plain_account.PlainAccount, error) {
	plainAcc, err := dataStorage.PlainAccs.GetPlainAccount(publicKeyHash)
	if err != nil {
		return nil, err
	}
	if plainAcc != nil {
		return plainAcc, nil
	}
	return dataStorage.CreatePlainAccount(publicKeyHash)
}

func (dataStorage *DataStorage) CreatePlainAccount(publicKeyHash []byte) (*plain_account.PlainAccount, error) {
	return dataStorage.PlainAccs.CreateNewPlainAccount(publicKeyHash)
}

func (dataStorage *DataStorage) AddStakePendingStake(publicKeyHash []byte, amount uint64, pendingType bool, blockHeight uint64) error {

	if pendingType {
		blockHeight += config_stake.GetPendingStakeWindow(blockHeight)
	} else {
		blockHeight += config_stake.GetPendingUnstakeWindow(blockHeight)
	}

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
		PublicKeyHash: publicKeyHash,
		PendingAmount: amount,
		PendingType:   pendingType,
	})

	return dataStorage.PendingStakes.Update(strconv.FormatUint(blockHeight, 10), pendingStakes)
}

func (dataStorage *DataStorage) ProcessPendingStakes(blockHeight uint64) error {

	pendingStakes, err := dataStorage.PendingStakes.GetPendingStakes(blockHeight)
	if err != nil {
		return err
	}

	if pendingStakes == nil {
		return nil
	}

	for _, pending := range pendingStakes.Pending {

		if pending.PendingType { //add

			var plainAcc *plain_account.PlainAccount
			if plainAcc, err = dataStorage.GetOrCreatePlainAccount(pending.PublicKeyHash); err != nil {
				return err
			}

			if err = plainAcc.AddStakeAvailable(true, pending.PendingAmount); err != nil {
				return err
			}

			if err = dataStorage.PlainAccs.Update(string(pending.PublicKeyHash), plainAcc); err != nil {
				return err
			}

		} else {

			var acc *account.Account
			var accs *accounts.Accounts

			if accs, acc, err = dataStorage.GetOrCreateAccount(config_coins.NATIVE_ASSET_FULL, pending.PublicKeyHash); err != nil {
				return err
			}

			if err = acc.AddBalance(true, pending.PendingAmount); err != nil {
				return err
			}

			if err = accs.Update(string(pending.PublicKeyHash), acc); err != nil {
				return err
			}

		}

	}

	dataStorage.PendingStakes.Delete(strconv.FormatUint(blockHeight, 10))
	return nil
}

func NewDataStorage(dbTx store_db_interface.StoreDBTransactionInterface) (out *DataStorage) {

	out = &DataStorage{
		dbTx,
		plain_accounts.NewPlainAccounts(dbTx),
		accounts.NewAccountsCollection(dbTx),
		pending_stakes_list.NewPendingStakesList(dbTx),
		assets.NewAssets(dbTx),
	}

	return
}
