package hash_map

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type HashMap struct {
	name           string
	Tx             store_db_interface.StoreDBTransactionInterface
	Count          uint64
	countCommitted uint64
	Changes        map[string]*ChangesMapElement
	changed        bool
	changesSize    map[string]*ChangesMapElement //used for computing the
	Committed      map[string]*CommittedMapElement
	keyLength      int
	CreateObject   func(key []byte, index uint64) (HashMapElementSerializableInterface, error)
	DeletedEvent   func([]byte) error
	StoredEvent    func([]byte, *CommittedMapElement) error
	Indexable      bool
}

func (hashMap *HashMap) deserialize(key, data []byte, index uint64) (HashMapElementSerializableInterface, error) {
	obj, err := hashMap.CreateObject(key, index)
	if err != nil {
		return nil, err
	}
	if err := obj.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return nil, err
	}
	return obj, nil
}

//support only for commited data
func (hashMap *HashMap) GetIndexByKey(key string) (uint64, error) {
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

//support only for commited data
func (hashMap *HashMap) GetKeyByIndex(index uint64) ([]byte, error) {
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

//support only for commited data
func (hashMap *HashMap) GetByIndex(index uint64) (data helpers.SerializableInterface, err error) {

	key, err := hashMap.GetKeyByIndex(index)
	if err != nil {
		return nil, err
	}

	return hashMap.Get(string(key))
}

//support only for commited data
func (hashMap *HashMap) GetRandom() (data helpers.SerializableInterface, err error) {
	if !hashMap.Indexable {
		return nil, errors.New("HashMap is not Indexable")
	}

	index := rand.Uint64() % hashMap.Count
	return hashMap.GetByIndex(index)
}

func (hashMap *HashMap) Get(key string) (out HashMapElementSerializableInterface, err error) {

	if hashMap.keyLength != 0 && len(key) != hashMap.keyLength {
		return nil, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element, nil
	}

	var outData []byte
	var index uint64

	if exists2 := hashMap.Committed[key]; exists2 != nil {
		if exists2.Element != nil {
			outData = helpers.CloneBytes(exists2.serialized)
			if hashMap.Indexable {
				index = exists2.Element.GetIndex()
			}
		}
	} else {
		//clone required because data could be altered afterwards
		outData = hashMap.Tx.Get(hashMap.name + ":map:" + key)
		if outData != nil && hashMap.Indexable {

			//safe because the bytes will be converted into an integer
			data := hashMap.Tx.Get(hashMap.name + ":listKeys:" + key)
			if data == nil {
				return nil, errors.New("Key not found")
			}

			if index, err = strconv.ParseUint(string(data), 10, 64); err != nil {
				return
			}
		}
	}

	if outData != nil {
		if out, err = hashMap.deserialize([]byte(key), outData, index); err != nil {
			return nil, err
		}
	}
	hashMap.Changes[key] = &ChangesMapElement{out, "view", 0, false}
	return
}

func (hashMap *HashMap) Exists(key string) (bool, error) {

	if hashMap.keyLength != 0 && len(key) != hashMap.keyLength {
		return false, errors.New("key length is invalid")
	}
	if exists := hashMap.Changes[key]; exists != nil {
		return exists.Element != nil, nil
	}
	if exists2 := hashMap.Committed[key]; exists2 != nil {
		return exists2.Element != nil, nil
	}

	return hashMap.Tx.Exists(hashMap.name + ":exists:" + key), nil
}

//this will verify if the data still exists
func (hashMap *HashMap) Create(key string, data HashMapElementSerializableInterface) error {
	exists, err := hashMap.Exists(key)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("Element already exists in Hashmap")
	}
	return hashMap.Update(key, data)
}

func (hashMap *HashMap) Update(key string, data HashMapElementSerializableInterface) error {

	if hashMap.keyLength != 0 && len(key) != hashMap.keyLength {
		return errors.New("key length is invalid")
	}
	if data == nil {
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
	if (exists != nil && exists.Element == nil) ||
		(exists == nil && !hashMap.Tx.Exists(hashMap.name+":exists:"+key)) {
		increase = true
	}

	if exists == nil {
		exists = new(ChangesMapElement)
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

func (hashMap *HashMap) Delete(key string) {

	exists := hashMap.Changes[key]

	decrease := false
	if (exists != nil && exists.Element != nil) ||
		(exists == nil && hashMap.Tx.Exists(hashMap.name+":exists:"+key)) {
		decrease = true
	}

	if exists == nil {
		exists = new(ChangesMapElement)
		hashMap.Changes[key] = exists
		hashMap.changesSize[key] = exists
	}
	exists.Status = "del"
	exists.Element = nil

	hashMap.changed = true

	if decrease {
		hashMap.Count -= 1

		if hashMap.Indexable {
			if exists.indexProcess {
				for _, v := range hashMap.Changes {
					if v.index > exists.index {
						v.index -= 1
						if v.Element != nil {
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

func (hashMap *HashMap) UpdateOrDelete(key string, data HashMapElementSerializableInterface) error {
	if data == nil {
		hashMap.Delete(key)
		return nil
	}
	return hashMap.Update(key, data)
}

func (hashMap *HashMap) ComputeChangesSize() (out uint64) {

	for k, v := range hashMap.changesSize {
		if v.Status == "update" {

			oldSize := 0

			serialized := helpers.SerializeToBytes(v.Element)
			newSize := len(serialized)

			isNew := false

			if exists := hashMap.Committed[k]; exists != nil {
				oldSize = exists.size
				if exists.Element == nil {
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

func (hashMap *HashMap) ResetChangesSize() {
	hashMap.changesSize = make(map[string]*ChangesMapElement)
}

func (hashMap *HashMap) CommitChanges() (err error) {

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
				committed = new(CommittedMapElement)
				hashMap.Committed[k] = committed
			}

			v.Status = "view"
			committed.Status = "view"
			committed.Element = nil
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
				committed = new(CommittedMapElement)
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
					if err = hashMap.StoredEvent([]byte(k), committed); err != nil {
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

func (hashMap *HashMap) SetTx(dbTx store_db_interface.StoreDBTransactionInterface) {
	hashMap.Tx = dbTx
}

func (hashMap *HashMap) Rollback() {
	hashMap.Changes = make(map[string]*ChangesMapElement)
	hashMap.changesSize = make(map[string]*ChangesMapElement)
	hashMap.Count = hashMap.countCommitted
	hashMap.changed = false
}

func (hashMap *HashMap) Reset() {
	hashMap.Committed = make(map[string]*CommittedMapElement)
	hashMap.Changes = make(map[string]*ChangesMapElement)
	hashMap.changesSize = make(map[string]*ChangesMapElement)
	hashMap.changed = false
}

func CreateNewHashMap(tx store_db_interface.StoreDBTransactionInterface, name string, keyLength int, indexable bool) (hashMap *HashMap) {

	if len(name) <= 4 {
		panic("Invalid name")
	}

	hashMap = &HashMap{
		name:        name,
		Committed:   make(map[string]*CommittedMapElement),
		Changes:     make(map[string]*ChangesMapElement),
		changesSize: make(map[string]*ChangesMapElement),
		Tx:          tx,
		Count:       0,
		keyLength:   keyLength,
		Indexable:   indexable,
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
