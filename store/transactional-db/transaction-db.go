package transactional_db

import (
	"errors"
	"sync"
	"sync/atomic"
)

type TransactionDB struct {
	database  *TransactionalDB
	changes   sync.Map
	committed sync.Map
	writeable bool
	index     uint64
	mutex     sync.Mutex
}

func (tx *TransactionDB) Put(key []byte, data []byte) error {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if !tx.writeable {
		return errors.New("DB is not writable")
	}
	keyStr := string(key)
	tx.changes.Store(keyStr, data)

	return nil
}

func (tx *TransactionDB) Get(key []byte) []byte {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	keyStr := string(key)

	value, exists := tx.changes.Load(keyStr)
	if exists {
		return value.([]byte)
	}
	value, exists = tx.committed.Load(keyStr)
	if exists {
		return value.([]byte)
	}

	data, exists := tx.database.temporary.Load(keyStr)
	if exists {
		temporary := data.(*TemporaryChanges)
		out, found := temporary.Get(tx.index)
		if found {
			return out
		}
	}

	return tx.database.db.Get(key)
}

func (tx *TransactionDB) Delete(key []byte) error {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	if !tx.writeable {
		return errors.New("DB is not writable")
	}
	keyStr := string(key)
	tx.changes.Delete(keyStr)

	return nil
}

func (tx *TransactionDB) Commit() {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	tx.changes.Range(func(key, value interface{}) bool {
		tx.committed.Store(key, value)
		return true
	})
	tx.changes = sync.Map{}
}

func (tx *TransactionDB) UnCommit() {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()
	tx.changes = sync.Map{}
}

func (tx *TransactionDB) Store() error {
	if !tx.writeable {
		return errors.New("DB is not writable")
	}
	tx.mutex.Lock()
	defer tx.mutex.Unlock()

	tx.committed.Range(func(key, value interface{}) bool {

		newTemporary := &TemporaryChanges{
			list: []*TemporaryChange{},
		}

		temporaryFound, _ := tx.database.temporary.LoadOrStore(key, newTemporary)
		temporary := temporaryFound.(*TemporaryChanges)

		before, found := temporary.Get(tx.index)
		if !found {
			before = tx.database.db.Get([]byte(key.(string)))
		}

		temporary.Insert(before, value.([]byte), tx.index)

		return true
	})

	tx.Close()
	return nil
}

func (tx TransactionDB) Close() {
	atomic.CompareAndSwapUint64(&tx.database.indexSmallestOpen, tx.index, tx.index+1)
	tx.writeable = false
}
