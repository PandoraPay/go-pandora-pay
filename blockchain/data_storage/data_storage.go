package data_storage

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/pending_future_list"
	"pandora-pay/blockchain/data_storage/pending_future_list/pending_future"
	"pandora-pay/blockchain/data_storage/pending_stakes_list"
	"pandora-pay/blockchain/data_storage/pending_stakes_list/pending_stakes"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/config/config_asset_fee"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type DataStorage struct {
	DBTx                       store_db_interface.StoreDBTransactionInterface
	Regs                       *registrations.Registrations
	PlainAccs                  *plain_accounts.PlainAccounts
	AccsCollection             *accounts.AccountsCollection
	PendingStakes              *pending_stakes_list.PendingStakesList
	PendingFutureCollection    *pending_future_list.PendingFutureCollection
	Asts                       *assets.Assets
	AstsFeeLiquidityCollection *assets.AssetsFeeLiquidityCollection
}

func (dataStorage *DataStorage) GetOrCreateAccount(assetId, publicKey []byte, validateRegistration bool) (*accounts.Accounts, *account.Account, error) {

	if validateRegistration {
		exists, err := dataStorage.Regs.Exists(string(publicKey))
		if err != nil {
			return nil, nil, err
		}
		if !exists {
			return nil, nil, errors.New("Can't create Account as it is not Registered")
		}
	}

	accs, err := dataStorage.AccsCollection.GetMap(assetId)
	if err != nil {
		return nil, nil, err
	}

	acc, err := accs.Get(string(publicKey))
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

func (dataStorage *DataStorage) CreateAccount(assetId, publicKey []byte, validateRegistration bool) (*accounts.Accounts, *account.Account, error) {

	if validateRegistration {
		exists, err := dataStorage.Regs.Exists(string(publicKey))
		if err != nil {
			return nil, nil, err
		}
		if !exists {
			return nil, nil, errors.New("Can't create Account as it is not Registered")
		}
	}

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

func (dataStorage *DataStorage) GetOrCreatePlainAccount(publicKey []byte, validateRegistration bool) (*plain_account.PlainAccount, error) {
	plainAcc, err := dataStorage.PlainAccs.Get(string(publicKey))
	if err != nil {
		return nil, err
	}
	if plainAcc != nil {
		return plainAcc, nil
	}
	return dataStorage.CreatePlainAccount(publicKey, validateRegistration)
}

func (dataStorage *DataStorage) CreatePlainAccount(publicKey []byte, validateRegistration bool) (*plain_account.PlainAccount, error) {

	if validateRegistration {
		exists, err := dataStorage.Regs.Exists(string(publicKey))
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("PlainAccount should not have been registered before")
		}
	}

	return dataStorage.PlainAccs.CreateNewPlainAccount(publicKey)
}

func (dataStorage *DataStorage) CreateRegistration(publicKey []byte, staked bool, spendPublicKey []byte) (*registration.Registration, error) {

	exists, err := dataStorage.PlainAccs.Exists(string(publicKey))
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("Can't register as a plain Account already exists")
	}

	return dataStorage.Regs.CreateNewRegistration(publicKey, staked, spendPublicKey)
}

func (dataStorage *DataStorage) AddPendingStake(publicKey []byte, amount *crypto.ElGamal, blockHeight uint64) error {

	reg, err := dataStorage.Regs.Get(string(publicKey))
	if err != nil {
		return err
	}

	if reg == nil {
		return errors.New("Account was not registered")
	}

	if !reg.Staked {
		return errors.New("reg.Staked is false")
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
		publicKey,
		amount.Serialize(),
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
		if acc, err = accs.Get(string(pending.PublicKey)); err != nil {
			return err
		}

		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		pendingAmount, err := new(crypto.ElGamal).Deserialize(pending.PendingAmount)
		if err != nil {
			return err
		}
		acc.Balance.AddEchanges(pendingAmount)

		if err = accs.Update(string(pending.PublicKey), acc); err != nil {
			return err
		}
	}

	dataStorage.PendingStakes.Delete(strconv.FormatUint(blockHeight, 10))
	return nil
}

func (dataStorage *DataStorage) AddPendingFuture(blockHeight uint64, txId []byte, payloadIndex byte, asset []byte, defaultResolution bool, parity bool, publicKeyList [][]byte, echangesAll []*crypto.ElGamal, multisigThreshold byte, multisigPublicKeys [][]byte) error {

	for i, publicKey := range publicKeyList {
		reg, err := dataStorage.Regs.Get(string(publicKey))
		if err != nil {
			return err
		}
		if reg == nil {
			return errors.New("Account was not registered")
		}
		if reg.Staked {
			return fmt.Errorf("reg.Staked should not be true for %d %s", i, publicKey)
		}
	}

	pendingFutureMap, err := dataStorage.PendingFutureCollection.GetMap(blockHeight)
	if err != nil {
		return err
	}

	key := string(txId) + "_" + strconv.Itoa(int(payloadIndex))

	pendingFuture, err := pendingFutureMap.Get(key)
	if err != nil {
		return err
	}

	if pendingFuture != nil {
		return errors.New("Already exists")
	}

	pendingFuture = pending_future.NewPendingFuture([]byte(key), 0, blockHeight)
	pendingFuture.TxId = txId
	pendingFuture.Asset = asset
	pendingFuture.DefaultResolution = defaultResolution
	pendingFuture.PayloadIndex = payloadIndex

	pendingFuture.ReceiverPublicKeys = make([][]byte, len(publicKeyList)/2)
	pendingFuture.ReceiverAmounts = make([][]byte, len(publicKeyList)/2)
	pendingFuture.SenderPublicKeys = make([][]byte, len(publicKeyList)/2)
	pendingFuture.SenderAmounts = make([][]byte, len(publicKeyList)/2)

	for i := range publicKeyList {
		if (i%2 == 0) == parity { //sender
			pendingFuture.SenderPublicKeys[i/2] = publicKeyList[i]
			pendingFuture.SenderAmounts[i/2] = echangesAll[i].Serialize()
		} else { //receiver
			pendingFuture.ReceiverPublicKeys[i/2] = publicKeyList[i]
			pendingFuture.ReceiverAmounts[i/2] = echangesAll[i].Serialize()
		}
	}

	pendingFuture.MultisigThreshold = multisigThreshold
	pendingFuture.MultisigPublicKeys = multisigPublicKeys

	return pendingFutureMap.Update(key, pendingFuture)
}

func (dataStorage *DataStorage) ProceedPendingFuture(resolution bool, pendingFuture *pending_future.PendingFuture) (err error) {

	if pendingFuture.Processed {
		return errors.New("pending Future already processed")
	}

	pendingFuture.Processed = true

	var acc *account.Account
	var pendingAmount *crypto.ElGamal

	accs, err := dataStorage.AccsCollection.GetMap(pendingFuture.Asset)
	if err != nil {
		return
	}

	for i := range pendingFuture.ReceiverPublicKeys {
		var key, amount []byte
		if resolution {
			key = pendingFuture.ReceiverPublicKeys[i]
			amount = pendingFuture.ReceiverAmounts[i]
		} else {
			key = pendingFuture.SenderPublicKeys[i]
			amount = pendingFuture.SenderAmounts[i]
		}

		if acc, err = accs.Get(string(key)); err != nil {
			return
		}
		if acc == nil {
			if acc, err = accs.CreateNewAccount(key); err != nil {
				return
			}
		}

		if pendingAmount, err = new(crypto.ElGamal).Deserialize(amount); err != nil {
			return
		}

		if !resolution {
			pendingAmount = pendingAmount.Neg()
		}

		acc.Balance.AddEchanges(pendingAmount) //neg is required to reverse
		if err = accs.Update(string(key), acc); err != nil {
			return
		}

	}

	return nil
}

func (dataStorage *DataStorage) ProcessPendingFuture(blockHeight uint64) error {

	pendingFutureMap, err := dataStorage.PendingFutureCollection.GetMap(blockHeight)
	if err != nil {
		return err
	}

	deleteKeys := make([]string, pendingFutureMap.Count)
	for i := uint64(0); i < pendingFutureMap.Count; i++ {

		pending, err := pendingFutureMap.GetByIndex(i)
		if err != nil {
			return err
		}

		deleteKeys[i] = string(pending.TxId) + "_" + strconv.Itoa(int(pending.PayloadIndex))

		if !pending.Processed {
			if err = dataStorage.ProceedPendingFuture(pending.DefaultResolution, pending); err != nil {
				return err
			}
		}

	}

	for _, key := range deleteKeys {
		pendingFutureMap.Delete(key)
	}

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

	return dataStorage.PlainAccs.Get(string(key))
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
		registrations.NewRegistrations(dbTx),
		plain_accounts.NewPlainAccounts(dbTx),
		accounts.NewAccountsCollection(dbTx),
		pending_stakes_list.NewPendingStakesList(dbTx),
		pending_future_list.NewPendingFutureCollection(dbTx),
		assets.NewAssets(dbTx),
		assets.NewAssetsFeeLiquidityCollection(dbTx),
	}

	return
}
