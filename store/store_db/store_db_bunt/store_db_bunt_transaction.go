package store_db_bunt

import (
	"errors"
	buntdb "github.com/tidwall/buntdb"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
)

type StoreDBBuntTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	buntTx *buntdb.Tx
	write  bool
}

func (tx *StoreDBBuntTransaction) IsWritable() bool {
	return tx.write
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
	//TODO: check if cloneBytes is necessary for BuntDB
	return helpers.CloneBytes(tx.Get(key))
}

func (tx *StoreDBBuntTransaction) PutClone(key string, value []byte) error {
	return tx.Put(key, helpers.CloneBytes(value))
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
