package transactional_db

import "sync"

type TemporaryChange struct {
	index  uint64
	before []byte
	after  []byte
}

type TemporaryChanges struct {
	database     *TransactionalDB
	list         []*TemporaryChange
	sync.RWMutex `json:"-"`
}

func (changes *TemporaryChanges) Insert(before, after []byte, index uint64, lock bool) {

	if lock {
		changes.Lock()
		defer changes.Unlock()
	}

	changes.list = append(changes.list, &TemporaryChange{index, before, after})
}

func (changes *TemporaryChanges) Get(index uint64, lock bool) ([]byte, bool) {

	if lock {
		changes.RLock()
		defer changes.RUnlock()
	}

	var found *TemporaryChange
	for _, change := range changes.list {
		if change.index < index && (found == nil || found.index < change.index) {
			found = change
		}
	}
	if found != nil {
		return found.after, true
	}

	return nil, false
}
