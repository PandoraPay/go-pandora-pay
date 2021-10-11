package store_db_bunt

import (
	"github.com/tidwall/buntdb"
	"os"
	store_db_interface "pandora-pay/store/store_db/store_db_interface"
)

const dbName = "bunt"

type StoreDBBunt struct {
	store_db_interface.StoreDBInterface
	DB   *buntdb.DB
	Name []byte
}

func (store *StoreDBBunt) Close() error {
	return store.DB.Close()
}

func (store *StoreDBBunt) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	return store.DB.View(func(buntTx *buntdb.Tx) error {
		tx := &StoreDBBuntTransaction{
			buntTx: buntTx,
		}
		return callback(tx)
	})
}

func (store *StoreDBBunt) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	return store.DB.Update(func(buntTx *buntdb.Tx) error {
		tx := &StoreDBBuntTransaction{
			buntTx: buntTx,
			write:  true,
		}
		return callback(tx)
	})
}

func CreateStoreDBBunt(name string, inMemory bool) (*StoreDBBunt, error) {

	var err error

	var prefix string
	if !inMemory {
		prefix = "./store"
		if _, err = os.Stat(prefix); os.IsNotExist(err) {
			if err = os.Mkdir(prefix, 0755); err != nil {
				return nil, err
			}
		}
		prefix += name + "_store" + "." + dbName
	} else {
		prefix = ":memory:"
	}

	store := &StoreDBBunt{
		Name: []byte(name),
	}

	// Open the my.store data file in your current directory.
	// It will be created if it doesn't exist.
	if store.DB, err = buntdb.Open(prefix); err != nil {
		return nil, err
	}

	return store, nil
}
