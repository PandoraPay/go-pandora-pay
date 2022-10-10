package hash_map

import (
	"pandora-pay/helpers/advanced_buffers"
	"pandora-pay/store/store_db/store_db_interface"
)

type HashMapInterface interface {
	ResetChangesSize()
	ComputeChangesSize() uint64
	CommitChanges() error
	Rollback()
	SetTx(tx store_db_interface.StoreDBTransactionInterface)
	WriteTransitionalChangesToStore(prefix string) (bool, error)
	DeleteTransitionalChangesFromStore(prefix string)
	ReadTransitionalChangesFromStore(prefix string) error
}

type HashMapElementSerializableInterface interface {
	comparable
	Validate() error
	Serialize(w *advanced_buffers.BufferWriter)
	Deserialize(r *advanced_buffers.BufferReader) error
	SetIndex(index uint64)
	SetKey(key []byte)
	GetIndex() uint64
	IsDeletable() bool
}

type ChangesMapElement[T HashMapElementSerializableInterface] struct {
	Element      T
	Status       string
	index        uint64
	indexProcess bool
}

type CommittedMapElement[T HashMapElementSerializableInterface] struct {
	Element    T
	Status     string
	Stored     string
	serialized []byte
	size       int
}
