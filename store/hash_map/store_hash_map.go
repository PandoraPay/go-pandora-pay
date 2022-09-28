package hash_map

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type HashMap[T HashMapElementSerializableInterface] struct {
	HashMapInterface
	name           string
	Tx             store_db_interface.StoreDBTransactionInterface
	Count          uint64
	countCommitted uint64
	Changes        map[string]*ChangesMapElement[T]
	changed        bool
	changesSize    map[string]*ChangesMapElement[T] //used for computing the
	Committed      map[string]*CommittedMapElement[T]
	keyLength      int
	CreateObject   func(key []byte, index uint64) (T, error)
	DeletedEvent   func(key []byte) error
	StoredEvent    func(key []byte, committed *CommittedMapElement[T], index uint64) error
	Indexable      bool
}

func (hashMap *HashMap[T]) deserialize(key, data []byte, index uint64) (T, error) {
	out, err := hashMap.CreateObject(key, index)
	if err != nil {
		return generics.Zero[T](), err
	}
	if err = out.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return generics.Zero[T](), err
	}
	return out, nil
}

// support only for commited data
func (hashMap *HashMap[T]) GetIndexByKey(key string) (uint64, error) {
	if !hashMap.Indexable {
		return 0, errors.New("HashMap is not Indexable")
	}

	if hashMap.changed {
		return 0, errors.New("GetIndexByKey is supported only when is committed")
	}

	//safe to Get because it won't change
	data := hashMap.Tx.Get(hashMap.name + ":listKeys:" + key)
	if data == nil {
		return 0, errors.New("Key not found")
	}

	return strconv.ParseUint(string(data), 10, 64)
}

// support only for commited data
func (hashMap *HashMap[T]) GetKeyByIndex(index uint64) ([]byte, error) {
	if !hashMap.Indexable {
		return nil, errors.New("HashMap is not Indexable")
	}

	if hashMap.changed {
		return nil, errors.New("GetIndexByKey is supported only when is committed")
	}

	if index >= hashMap.Count {
		return nil, errors.New("Index exceeds count")
	}

	//Clone require because key might get altered afterwards
	key := hashMap.Tx.Get(hashMap.name + ":list:" + strconv.FormatUint(index, 10))
	if key == nil {
		return nil, errors.New("Not found")
	}

	return key, nil
}

// support only for commited data
func (hashMap *HashMap[T]) GetByIndex(index uint64) (T, error) {

	key, err := hashMap.GetKeyByIndex(index)
	if err != nil {
		return generics.Zero[T](), err
	}

	return hashMap.Get(string(key))
}

// support only for commited data
func (hashMap *HashMap[T]) GetRandom() (T, error) {

	if !hashMap.Indexable {
		return generics.Zero[T](), errors.New("HashMap is not Indexable")
	}

	index := rand.Uint64() % hashMap.Count
	return hashMap.GetByIndex(index)
}

func (hashMap *HashMap[T]) Get(key string) (out T, err error) {

	if hashMap.keyLength != 0 && len(key) != hashMap.keyLength {
		return generics.Zero[T](), errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element, nil
	}

	var outData []byte
	var index uint64

	if exists := hashMap.Committed[key]; exists != nil {
		if !generics.IsZero(exists.Element) {
			outData = helpers.CloneBytes(exists.serialized)
			if hashMap.Indexable {
				index = exists.Element.GetIndex()
			}
		}
	} else {
		//clone required because data could be altered afterwards
		outData = hashMap.Tx.Get(hashMap.name + ":map:" + key)
		if outData != nil && hashMap.Indexable {

			//safe because the bytes will be converted into an integer
			data := hashMap.Tx.Get(hashMap.name + ":listKeys:" + key)
			if data == nil {
				return generics.Zero[T](), errors.New("Key not found")
			}

			if index, err = strconv.ParseUint(string(data), 10, 64); err != nil {
				return
			}
		}
	}

	if outData != nil {
		if out, err = hashMap.deserialize([]byte(key), outData, index); err != nil {
			return generics.Zero[T](), err
		}
	}
	hashMap.Changes[key] = &ChangesMapElement[T]{out, "view", 0, false}
	return
}

func (hashMap *HashMap[T]) Exists(key string) (bool, error) {

	if hashMap.keyLength != 0 && len(key) != hashMap.keyLength {
		return false, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return !generics.IsZero(exists.Element), nil
	}
	if exists := hashMap.Committed[key]; exists != nil {
		return !generics.IsZero(exists.Element), nil
	}

	return hashMap.Tx.Exists(hashMap.name + ":exists:" + key), nil
}

// this will verify if the data still exists
func (hashMap *HashMap[T]) Create(key string, data T) error {
	exists, err := hashMap.Exists(key)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("Element already exists in Hashmap")
	}
	return hashMap.Update(key, data)
}

func (hashMap *HashMap[T]) Update(key string, data T) error {

	if hashMap.keyLength != 0 && len(key) != hashMap.keyLength {
		return errors.New("key length is invalid")
	}
	if generics.IsZero(data) {
		return errors.New("Data is null and it should not be")
	}

	data.SetKey([]byte(key))

	if err := data.Validate(); err != nil {
		return err
	}

	if data.IsDeletable() {
		hashMap.Delete(key)
		return nil
	}

	exists := hashMap.Changes[key]

	increase := false
	if (exists != nil && generics.IsZero(exists.Element)) ||
		(exists == nil && !hashMap.Tx.Exists(hashMap.name+":exists:"+key)) {
		increase = true
	}

	if exists == nil {
		exists = new(ChangesMapElement[T])
		hashMap.Changes[key] = exists
		hashMap.changesSize[key] = exists
	}
	exists.Status = "update"
	exists.Element = data

	hashMap.changed = true

	if increase {
		if hashMap.Indexable {
			exists.index = hashMap.Count
			exists.Element.SetIndex(hashMap.Count)
			exists.indexProcess = true
		}
		hashMap.Count += 1
	}

	return nil
}

func (hashMap *HashMap[T]) Delete(key string) {

	exists := hashMap.Changes[key]

	decrease := false
	if (exists != nil && !generics.IsZero[T](exists.Element)) ||
		(exists == nil && hashMap.Tx.Exists(hashMap.name+":exists:"+key)) {
		decrease = true
	}

	if exists == nil {
		exists = new(ChangesMapElement[T])
		hashMap.Changes[key] = exists
		hashMap.changesSize[key] = exists
	}
	exists.Status = "del"
	exists.Element = generics.Zero[T]()

	hashMap.changed = true

	if decrease {
		hashMap.Count -= 1

		if hashMap.Indexable {
			if exists.indexProcess {
				for _, v := range hashMap.Changes {
					if v.index > exists.index {
						v.index -= 1
						if !generics.IsZero(v.Element) {
							v.Element.SetIndex(v.index)
						}
					}
				}
				exists.indexProcess = false
			} else {
				exists.index = hashMap.Count
				exists.indexProcess = true
			}
		}

	}

	return
}

func (hashMap *HashMap[T]) UpdateOrDelete(key string, data T) error {
	if generics.IsZero(data) {
		hashMap.Delete(key)
		return nil
	}
	return hashMap.Update(key, data)
}

func (hashMap *HashMap[T]) ComputeChangesSize() (out uint64) {

	for k, v := range hashMap.changesSize {
		if v.Status == "update" {

			oldSize := 0

			serialized := helpers.SerializeToBytes(v.Element)
			newSize := len(serialized)

			isNew := false

			if exists := hashMap.Committed[k]; exists != nil {
				oldSize = exists.size
				if generics.IsZero(exists.Element) {
					isNew = true
				}
			} else {
				//safe because only length is used
				if data := hashMap.Tx.Get(hashMap.name + ":map:" + k); data != nil {
					oldSize = len(data)
				} else {
					isNew = true
				}
			}

			if newSize > oldSize {
				out += uint64(newSize - oldSize)
				if isNew {
					out += uint64(len(k))
				}
			}

		}
	}

	return
}

func (hashMap *HashMap[T]) ResetChangesSize() {
	hashMap.changesSize = make(map[string]*ChangesMapElement[T])
}

func (hashMap *HashMap[T]) CommitChanges() (err error) {

	removed := make([]string, len(hashMap.Changes))

	c := 0
	for k, v := range hashMap.Changes {
		if hashMap.keyLength != 0 && len(k) != hashMap.keyLength {
			return errors.New("key length is invalid")
		}
		if v.Status == "update" {
			removed[c] = k
			c += 1
		}
	}

	for k, v := range hashMap.Changes {

		if v.Status == "del" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommittedMapElement[T])
				hashMap.Committed[k] = committed
			}

			v.Status = "view"
			committed.Status = "view"
			committed.Element = generics.Zero[T]()
			committed.size = 0
			committed.serialized = nil

			if hashMap.Tx.Exists(hashMap.name + ":exists:" + k) {

				if hashMap.Tx.IsWritable() {

					hashMap.Tx.Delete(hashMap.name + ":map:" + k)
					hashMap.Tx.Delete(hashMap.name + ":exists:" + k)

					if hashMap.Indexable && v.indexProcess {
						hashMap.Tx.Delete(hashMap.name + ":list:" + strconv.FormatUint(v.index, 10))
						hashMap.Tx.Delete(hashMap.name + ":listKeys:" + k)
					}

				}

				if hashMap.DeletedEvent != nil {
					if err = hashMap.DeletedEvent([]byte(k)); err != nil {
						return
					}
				}

				committed.Stored = "del"
			} else {
				committed.Stored = "view"
			}

			v.indexProcess = false
		}

	}

	for k, v := range hashMap.Changes {

		if v.Status == "update" {

			committed := hashMap.Committed[k]
			if committed == nil {
				committed = new(CommittedMapElement[T])
				hashMap.Committed[k] = committed
			}

			committed.Element = v.Element

			if !hashMap.Tx.Exists(hashMap.name + ":exists:" + k) {

				if hashMap.Tx.IsWritable() {

					//safe
					hashMap.Tx.Put(hashMap.name+":exists:"+k, []byte{1})

					if hashMap.Indexable && v.indexProcess {
						//safe
						hashMap.Tx.Put(hashMap.name+":list:"+strconv.FormatUint(v.index, 10), []byte(k))
						//safe
						hashMap.Tx.Put(hashMap.name+":listKeys:"+k, []byte(strconv.FormatUint(v.index, 10)))
					}

				}

				if hashMap.StoredEvent != nil {
					if err = hashMap.StoredEvent([]byte(k), committed, v.index); err != nil {
						return
					}
				}
			}

			committed.serialized = helpers.SerializeToBytes(v.Element)
			committed.size = len(committed.serialized)

			if hashMap.Tx.IsWritable() {
				//clone required because the element could change later on
				hashMap.Tx.Put(hashMap.name+":map:"+k, committed.serialized)
			}

			committed.Status = "view"
			committed.Stored = "update"

			v.indexProcess = false
		}
	}

	for i := 0; i < c; i++ {
		delete(hashMap.Changes, removed[i])
	}

	hashMap.countCommitted = hashMap.Count

	if hashMap.Tx.IsWritable() {
		buf := make([]byte, binary.MaxVarintLen64)
		n := binary.PutUvarint(buf, hashMap.Count)
		//safe
		hashMap.Tx.Put(hashMap.name+":count", buf[:n])
	}

	hashMap.changed = false

	return
}

func (hashMap *HashMap[T]) SetTx(dbTx store_db_interface.StoreDBTransactionInterface) {
	hashMap.Tx = dbTx
}

func (hashMap *HashMap[T]) Rollback() {
	hashMap.Changes = make(map[string]*ChangesMapElement[T])
	hashMap.changesSize = make(map[string]*ChangesMapElement[T])
	hashMap.Count = hashMap.countCommitted
	hashMap.changed = false
}

func (hashMap *HashMap[T]) Reset() {
	hashMap.Committed = make(map[string]*CommittedMapElement[T])
	hashMap.Changes = make(map[string]*ChangesMapElement[T])
	hashMap.changesSize = make(map[string]*ChangesMapElement[T])
	hashMap.changed = false
}

func CreateNewHashMap[T HashMapElementSerializableInterface](tx store_db_interface.StoreDBTransactionInterface, name string, keyLength int, indexable bool) (hashMap *HashMap[T]) {

	if len(name) <= 4 {
		panic("Invalid name")
	}

	hashMap = &HashMap[T]{
		nil,
		name,
		tx,
		0,
		0,
		make(map[string]*ChangesMapElement[T]),
		false,
		make(map[string]*ChangesMapElement[T]),
		make(map[string]*CommittedMapElement[T]),
		keyLength,
		nil,
		nil,
		nil,
		indexable,
	}

	//safe to Get because data will be converted into an integer
	if buffer := tx.Get(hashMap.name + ":count"); buffer != nil {
		count, p := binary.Uvarint(buffer)
		if p <= 0 {
			panic("Error reading")
		}
		hashMap.Count = count
	}
	hashMap.countCommitted = hashMap.Count

	return
}
