package store_db_bunt

import (
	buntdb "github.com/tidwall/buntdb"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBBuntTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	buntTx *buntdb.Tx
}

func (tx *StoreDBBuntTransaction) Put(key []byte, value []byte) (err error) {
	_, _, err = tx.buntTx.Set(string(key), string(value), nil)
	return
}

func (tx *StoreDBBuntTransaction) Get(key []byte) (out []byte) {
	data, err := tx.buntTx.Get(string(key), false)
	if err == nil {
		out = []byte(data)
	}
	return
}

func (tx *StoreDBBuntTransaction) Delete(key []byte) (err error) {
	_, err = tx.buntTx.Delete(string(key))
	return err
}
