package transactional_db

import (
	kv_db "pandora-pay/store/kv-db"
	memory_db "pandora-pay/store/memory-db"
	"sync"
	"sync/atomic"
	"time"
)

type TransactionalDB struct {
	db    kv_db.KeyValueDB
	index uint64

	temporary sync.Map //it stores []*TemporaryChanges
	indexes   sync.Map //it stores []uint64

	sync.RWMutex `json:"-"`
}

func (db *TransactionalDB) garbageCollector() {

	var lastRemovedIndex = uint64(0)
	for {

		_, exists := db.indexes.Load(lastRemovedIndex)
		if !exists {
			lastRemovedIndex += 1

			db.temporary.Range(func(key, value interface{}) bool {
				temporary := value.(*TemporaryChanges)

				var found bool
				temporary.RLock()
				for _, temporaryChange := range temporary.list {
					if temporaryChange.index <= lastRemovedIndex {
						found = true
						break
					}
				}
				temporary.RUnlock()
				if found {
					temporary.Lock()
					for i, temporaryChange := range temporary.list {
						if temporaryChange.index == lastRemovedIndex {

							if temporaryChange.after == nil {
								db.db.Delete([]byte(key.(string)))
							} else {
								db.db.Put([]byte(key.(string)), temporaryChange.after)
							}

							temporary.list[i] = temporary.list[len(temporary.list)-1]
							temporary.list = temporary.list[:len(temporary.list)-1]
						}
					}
					temporary.Unlock()
				}
				return true
			})

		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (db *TransactionalDB) getNewIndex() uint64 {
	for {
		index := atomic.AddUint64(&db.index, 1)
		_, exists := db.indexes.LoadOrStore(index, time.Now().Unix())
		if !exists {
			return index
		}
	}
}

func (db *TransactionalDB) createTx(writeable bool) *TransactionDB {
	return &TransactionDB{
		database:  db,
		index:     db.getNewIndex(),
		writeable: writeable,
	}
}

func (db *TransactionalDB) View() *TransactionDB {
	return db.createTx(false)
}

func (db *TransactionalDB) Update() *TransactionDB {
	return db.createTx(true)
}

func CreateTransactionalDB() *TransactionalDB {

	database := memory_db.MemoryStoreCreate()

	db := &TransactionalDB{
		db:        kv_db.KeyValueDB(database),
		temporary: sync.Map{},
		indexes:   sync.Map{},
		index:     0,
	}

	go db.garbageCollector()

	return db
}
