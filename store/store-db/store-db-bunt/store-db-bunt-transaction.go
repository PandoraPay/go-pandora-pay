package store_db_bunt

import (
	"errors"
	buntdb "github.com/tidwall/buntdb"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBBuntTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	buntTx *buntdb.Tx
	write  bool
}

func (tx *StoreDBBuntTransaction) Put(key string, value []byte) (err error) {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
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

func (tx *StoreDBBuntTransaction) Exists(key string) bool {
	_, err := tx.buntTx.Get(key, false)
	if err == nil {
		return true
	}
	return false
}

func (tx *StoreDBBuntTransaction) GetClone(key string) (out []byte) {
	return tx.Get(key) //not required
}

func (tx *StoreDBBuntTransaction) Delete(key string) (err error) {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	_, err = tx.buntTx.Delete(key)
	return err
}

func (tx *StoreDBBuntTransaction) DeleteForcefully(key string) (err error) {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	_, err = tx.buntTx.Delete(key)
	if err == buntdb.ErrNotFound {
		return nil
	}
	return err
}
