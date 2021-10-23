package data_storage

import (
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/store/store_db/store_db_interface"
)

type DataStorage struct {
	Tx             store_db_interface.StoreDBTransactionInterface
	Regs           *registrations.Registrations
	PlainAccs      *plain_accounts.PlainAccounts
	AccsCollection *accounts.AccountsCollection
	Asts           *assets.Assets
}

func (data *DataStorage) CommitChanges() (err error) {
	if err = data.AccsCollection.CommitChanges(); err != nil {
		return
	}
	if err = data.Asts.CommitChanges(); err != nil {
		return
	}
	if err = data.Regs.CommitChanges(); err != nil {
		return
	}
	return data.PlainAccs.CommitChanges()
}

func (data *DataStorage) Rollback() {
	data.AccsCollection.Rollback()
	data.Asts.Rollback()
	data.Regs.Rollback()
	data.PlainAccs.Rollback()
}

func (data *DataStorage) CloneCommitted() (err error) {
	if err = data.AccsCollection.CloneCommitted(); err != nil {
		return
	}
	if err = data.Asts.CloneCommitted(); err != nil {
		return
	}
	if err = data.Regs.CloneCommitted(); err != nil {
		return
	}
	return data.PlainAccs.CloneCommitted()
}

func (data *DataStorage) SetTx(dbTx store_db_interface.StoreDBTransactionInterface) {
	data.Tx = dbTx
	data.AccsCollection.SetTx(dbTx)
	data.Asts.SetTx(dbTx)
	data.PlainAccs.SetTx(dbTx)
	data.Regs.SetTx(dbTx)
}

func (data *DataStorage) ReadTransitionalChangesFromStore(prefix string) (err error) {
	if err = data.AccsCollection.ReadTransitionalChangesFromStore(prefix); err != nil {
		return
	}
	if err = data.PlainAccs.ReadTransitionalChangesFromStore(prefix); err != nil {
		return
	}
	if err = data.Asts.ReadTransitionalChangesFromStore(prefix); err != nil {
		return
	}
	return data.Regs.ReadTransitionalChangesFromStore(prefix)
}

func (data *DataStorage) WriteTransitionalChangesToStore(prefix string) (err error) {
	if err = data.AccsCollection.WriteTransitionalChangesToStore(prefix); err != nil {
		return
	}
	if err = data.PlainAccs.WriteTransitionalChangesToStore(prefix); err != nil {
		return
	}
	if err = data.Asts.WriteTransitionalChangesToStore(prefix); err != nil {
		return
	}
	return data.Regs.WriteTransitionalChangesToStore(prefix)
}
func (data *DataStorage) DeleteTransitionalChangesFromStore(prefix string) (err error) {
	if err = data.AccsCollection.DeleteTransitionalChangesFromStore(prefix); err != nil {
		return
	}
	if err = data.PlainAccs.DeleteTransitionalChangesFromStore(prefix); err != nil {
		return
	}
	if err = data.Asts.DeleteTransitionalChangesFromStore(prefix); err != nil {
		return
	}
	return data.Regs.DeleteTransitionalChangesFromStore(prefix)
}

func CreateDataStorage(tx store_db_interface.StoreDBTransactionInterface) *DataStorage {
	return &DataStorage{
		tx,
		registrations.NewRegistrations(tx),
		plain_accounts.NewPlainAccounts(tx),
		accounts.NewAccountsCollection(tx),
		assets.NewAssets(tx),
	}
}