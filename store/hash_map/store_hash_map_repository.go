package hash_map

import "pandora-pay/store/store_db/store_db_interface"

type StoreHashMapRepository struct {
	GetList func() []*HashMap
}

func (repository *StoreHashMapRepository) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	list := repository.GetList()
	for _, it := range list {
		it.SetTx(tx)
	}
}

func (repository *StoreHashMapRepository) Rollback() {
	list := repository.GetList()
	for _, it := range list {
		it.Rollback()
	}
}

func (repository *StoreHashMapRepository) CloneCommitted() (err error) {
	list := repository.GetList()
	for _, it := range list {
		if err = it.CloneCommitted(); err != nil {
			return
		}
	}
	return
}

func (repository *StoreHashMapRepository) CommitChanges() (err error) {
	list := repository.GetList()
	for _, it := range list {
		if err = it.CommitChanges(); err != nil {
			return
		}
	}
	return
}

func (repository *StoreHashMapRepository) WriteTransitionalChangesToStore(prefix string) (err error) {
	list := repository.GetList()
	for _, it := range list {
		if err = it.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
	}
	return
}

func (repository *StoreHashMapRepository) ReadTransitionalChangesFromStore(prefix string) (err error) {
	list := repository.GetList()
	for _, it := range list {
		if err = it.ReadTransitionalChangesFromStore(prefix); err != nil {
			return
		}
	}
	return
}
func (repository *StoreHashMapRepository) DeleteTransitionalChangesFromStore(prefix string) {
	list := repository.GetList()
	for _, it := range list {
		it.DeleteTransitionalChangesFromStore(prefix)
	}
	return
}
