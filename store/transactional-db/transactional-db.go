package transactional_db

import (
	kv_db "pandora-pay/store/kv-db"
	memory_db "pandora-pay/store/memory-db"
	"sync"
	"sync/atomic"
)

type TransactionalDB struct {
	db                kv_db.KeyValueDB
	index             uint64
	indexSmallestOpen uint64

	temporary sync.Map //it stores []

	sync.RWMutex `json:"-"`
}

func (db *TransactionalDB) View() *TransactionDB {
	index := atomic.AddUint64(&db.index, 1)
	return &TransactionDB{
		database: db,
		index:    index,
	}
}

func (db *TransactionalDB) Update() *TransactionDB {
	index := atomic.AddUint64(&db.index, 1)
	return &TransactionDB{
		database:  db,
		writeable: true,
		index:     index,
	}
}

func CreateTransactionalDB() *TransactionalDB {

	database := memory_db.MemoryStoreCreate()

	return &TransactionalDB{
		db:        kv_db.KeyValueDB(database),
		temporary: sync.Map{},
	}
}
