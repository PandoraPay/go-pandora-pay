package kv_db

type KeyValueDB interface {
	Put(key []byte, data []byte)
	Get(key []byte) []byte
	Delete(key []byte)
}
