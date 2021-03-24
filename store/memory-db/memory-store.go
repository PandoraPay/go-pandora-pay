package memory_db

type MemoryStore struct {
	data map[string][]byte
}

func (store *MemoryStore) Put(key []byte, data []byte) {
	store.data[string(key)] = data
}

func (store *MemoryStore) Get(key []byte) []byte {
	return store.data[string(key)]
}

func (store *MemoryStore) Delete(key []byte) {
	delete(store.data, string(key))
}

func MemoryStoreCreate() *MemoryStore {
	return &MemoryStore{
		data: make(map[string][]byte),
	}
}
