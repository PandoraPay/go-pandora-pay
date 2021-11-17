package hash_map

import "pandora-pay/helpers"

type HashMapElementSerializableInterface interface {
	helpers.SerializableInterface
	SetIndex(index uint64)
	SetKey(key []byte)
	GetIndex() uint64
}

type ChangesMapElement struct {
	Element      HashMapElementSerializableInterface
	Status       string
	index        uint64
	indexProcess bool
}

type CommittedMapElement struct {
	Element    HashMapElementSerializableInterface
	Status     string
	Stored     string
	serialized []byte
	size       int
}
