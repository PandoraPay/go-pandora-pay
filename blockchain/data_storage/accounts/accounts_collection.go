package accounts

import (
	"errors"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type AccountsCollection struct {
	tx      store_db_interface.StoreDBTransactionInterface
	accsMap map[string]*Accounts
}

func (collection *AccountsCollection) GetAllMap() map[string]*Accounts {
	return collection.accsMap
}

func (collection *AccountsCollection) GetAccountAssetsCount(key []byte) (uint64, error) {

	var count uint64
	var err error

	data := collection.tx.Get("accounts:assetsCount:" + string(key))
	if data != nil {
		if count, err = helpers.NewBufferReader(data).ReadUvarint(); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (collection *AccountsCollection) GetMap(assetId []byte) (*Accounts, error) {

	if len(assetId) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Asset was not found")
	}

	accs := collection.accsMap[string(assetId)]
	if accs == nil {
		var err error
		if accs, err = NewAccounts(collection.tx, assetId); err != nil {
			return nil, err
		}
		collection.accsMap[string(assetId)] = accs
	}
	return accs, nil
}

func (collection *AccountsCollection) GetAccountAssets(key []byte) ([][]byte, error) {

	count, err := collection.GetAccountAssetsCount(key)
	if err != nil {
		return nil, err
	}

	out := make([][]byte, count)

	for i := uint64(0); i < count; i++ {
		assetId := collection.tx.Get("accounts:assetByIndex:" + string(key) + ":" + strconv.FormatUint(i, 10))
		if assetId == nil {
			return nil, errors.New("Error reading AssetId")
		}
		out[i] = assetId
	}

	return out, nil
}

func (collection *AccountsCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
	for _, accs := range collection.accsMap {
		accs.SetTx(tx)
	}
}

func (collection *AccountsCollection) Rollback() {
	for _, accs := range collection.accsMap {
		accs.Rollback()
	}
}

func (collection *AccountsCollection) CloneCommitted() (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.CloneCommitted(); err != nil {
			return
		}
	}
	return
}

func (collection *AccountsCollection) CommitChanges() (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.CommitChanges(); err != nil {
			return
		}
	}
	return
}

func (collection *AccountsCollection) WriteTransitionalChangesToStore(prefix string) (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
	}
	return
}

func (collection *AccountsCollection) ReadTransitionalChangesFromStore(prefix string) (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.ReadTransitionalChangesFromStore(prefix); err != nil {
			return
		}
	}
	return
}
func (collection *AccountsCollection) DeleteTransitionalChangesFromStore(prefix string) (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.DeleteTransitionalChangesFromStore(prefix); err != nil {
			return
		}
	}
	return
}

func NewAccountsCollection(tx store_db_interface.StoreDBTransactionInterface) *AccountsCollection {
	return &AccountsCollection{
		tx,
		make(map[string]*Accounts),
	}
}
