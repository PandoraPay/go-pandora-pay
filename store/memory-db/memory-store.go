package memory_db

import "sync"

type MemoryStore struct {
	data         map[string][]byte
	sync.RWMutex `json:"-"`
}

func (store *MemoryStore) Put(key []byte, data []byte) {
	store.Lock()
	store.data[string(key)] = data
	store.Unlock()
}

func (store *MemoryStore) Get(key []byte) []byte {
	store.RLock()
	defer store.RUnlock()
	return store.data[string(key)]
}

func (store *MemoryStore) Delete(key []byte) {
	store.Lock()
	delete(store.data, string(key))
	store.RUnlock()
}

func MemoryStoreCreate() *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}
