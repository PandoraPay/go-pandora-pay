package store_db_bunt

import (
	buntdb "github.com/tidwall/buntdb"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBBuntTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	buntTx *buntdb.Tx
}

func (tx *StoreDBBuntTransaction) Put(key string, value []byte) (err error) {
	_, _, err = tx.buntTx.Set(key, string(value), nil)
	return
}

func (tx *StoreDBBuntTransaction) Get(key string) (out []byte) {
	data, err := tx.buntTx.Get(key, false)
	if err == nil {
		out = []byte(data)
	}
	return
}

func (tx *StoreDBBuntTransaction) Delete(key string) (err error) {
	_, err = tx.buntTx.Delete(key)
	return err
}

func (tx *StoreDBBuntTransaction) DeleteForcefully(key string) (err error) {
	_, err = tx.buntTx.Delete(key)
	if err == buntdb.ErrNotFound {
		return nil
	}
	return err
}
