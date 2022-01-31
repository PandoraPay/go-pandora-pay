package hash_map

import "pandora-pay/store/store_db/store_db_interface"

type StoreHashMapRepository struct {
	GetList func(computeChangesSize bool) []*HashMap
}

func (repository *StoreHashMapRepository) SetTx(dbTx store_db_interface.StoreDBTransactionInterface) {
	list := repository.GetList(false)
	for _, it := range list {
		it.SetTx(dbTx)
	}
}

func (repository *StoreHashMapRepository) ComputeChangesSize() (out uint64) {
	list := repository.GetList(true)
	for _, it := range list {
		out += it.ComputeChangesSize()
	}
	return
}

func (repository *StoreHashMapRepository) ResetChangesSize() {
	list := repository.GetList(false)
	for _, it := range list {
		it.ResetChangesSize()
	}
}

func (repository *StoreHashMapRepository) Rollback() {
	list := repository.GetList(false)
	for _, it := range list {
		it.Rollback()
	}
}

func (repository *StoreHashMapRepository) CommitChanges() (err error) {
	list := repository.GetList(false)
	for _, it := range list {
		if err = it.CommitChanges(); err != nil {
			return
		}
	}
	return
}

func (repository *StoreHashMapRepository) WriteTransitionalChangesToStore(prefix string) (err error) {
	list := repository.GetList(false)
	for _, it := range list {
		if err = it.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
	}
	return
}

func (repository *StoreHashMapRepository) ReadTransitionalChangesFromStore(prefix string) (err error) {
	list := repository.GetList(false)
	for _, it := range list {
		if err = it.ReadTransitionalChangesFromStore(prefix); err != nil {
			return
		}
	}
	return
}
func (repository *StoreHashMapRepository) DeleteTransitionalChangesFromStore(prefix string) {
	list := repository.GetList(false)
	for _, it := range list {
		it.DeleteTransitionalChangesFromStore(prefix)
	}
	return
}
