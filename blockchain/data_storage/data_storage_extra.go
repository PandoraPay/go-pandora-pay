package data_storage

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type dataStorageTransitionCollectionsKeys struct {
	Accounts             [][]byte
	AssetsFeeLiquidities [][]byte
}

func (dataStorage *DataStorage) GetListWithoutCollections() (list []*hash_map.HashMap) {
	return []*hash_map.HashMap{
		dataStorage.Regs.HashMap,
		dataStorage.PlainAccs.HashMap,
		dataStorage.PendingStakes.HashMap,
		dataStorage.Asts.HashMap,
	}
}

func (dataStorage *DataStorage) GetList(computeChangesSize bool) (list []*hash_map.HashMap) {

	list = dataStorage.GetListWithoutCollections()
	list = append(list, dataStorage.AccsCollection.GetAllHashmaps()...)

	if !computeChangesSize {
		list = append(list, dataStorage.AstsFeeLiquidityCollection.GetAllHashmaps()...)
	}

	return
}

func (dataStorage *DataStorage) ComputeChangesSize() (out uint64) {
	list := dataStorage.GetList(true)
	for _, it := range list {
		out += it.ComputeChangesSize()
	}
	return
}

func (dataStorage *DataStorage) ResetChangesSize() {
	list := dataStorage.GetList(false)
	for _, it := range list {
		it.ResetChangesSize()
	}
}

func (dataStorage *DataStorage) Rollback() {
	list := dataStorage.GetList(false)
	for _, it := range list {
		it.Rollback()
	}
}

func (dataStorage *DataStorage) CommitChanges() (err error) {
	list := dataStorage.GetList(false)
	for _, it := range list {
		if err = it.CommitChanges(); err != nil {
			return
		}
	}
	return
}

func (dataStorage *DataStorage) SetTx(dbTx store_db_interface.StoreDBTransactionInterface) {
	dataStorage.DBTx = dbTx
	list := dataStorage.GetList(true)
	for _, it := range list {
		it.SetTx(dbTx)
	}
	dataStorage.AccsCollection.SetTx(dbTx)
	dataStorage.AstsFeeLiquidityCollection.SetTx(dbTx)
}

func (dataStorage *DataStorage) WriteTransitionalChangesToStore(prefix string) (err error) {

	list := dataStorage.GetListWithoutCollections()
	for _, it := range list {
		if _, err = it.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
	}

	transitionCollectionsKeys := &dataStorageTransitionCollectionsKeys{
		make([][]byte, 0),
		make([][]byte, 0),
	}

	accounts := dataStorage.AccsCollection.GetAllMaps()
	var hasData, hasData2 bool
	for _, it := range accounts {
		if hasData, err = it.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData {
			transitionCollectionsKeys.Accounts = append(transitionCollectionsKeys.Accounts, it.Asset)
		}
	}

	liquidityMaxHeaps := dataStorage.AstsFeeLiquidityCollection.GetAllMaps()
	for key, it := range liquidityMaxHeaps {
		if hasData, err = it.HashMap.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData2, err = it.DictMap.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData || hasData2 {
			transitionCollectionsKeys.AssetsFeeLiquidities = append(transitionCollectionsKeys.AssetsFeeLiquidities, []byte(key))
		}
	}

	bytes, err := msgpack.Marshal(transitionCollectionsKeys)
	if err != nil {
		return
	}

	dataStorage.DBTx.Put("dataStorage:transitionsCollectionsKeys:"+prefix, bytes)

	return
}

func (dataStorage *DataStorage) ReadTransitionalChangesFromStore(prefix string) error {

	list := dataStorage.GetListWithoutCollections()
	for _, it := range list {
		if err := it.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
	}

	bytes := dataStorage.DBTx.Get("dataStorage:transitionsCollectionsKeys:" + prefix)
	if bytes == nil {
		return errors.New("transitionsCollectionsKeys is null")
	}

	transitionCollectionsKeys := &dataStorageTransitionCollectionsKeys{}
	if err := msgpack.Unmarshal(bytes, transitionCollectionsKeys); err != nil {
		return err
	}

	for _, key := range transitionCollectionsKeys.Accounts {
		accs, err := dataStorage.AccsCollection.GetMap(key)
		if err != nil {
			return err
		}
		if err := accs.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
	}

	for _, key := range transitionCollectionsKeys.AssetsFeeLiquidities {
		maxHeap, err := dataStorage.AstsFeeLiquidityCollection.GetMaxHeap(key)
		if err != nil {
			return err
		}
		if err := maxHeap.HashMap.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
		if err := maxHeap.DictMap.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
	}

	return nil
}
func (dataStorage *DataStorage) DeleteTransitionalChangesFromStore(prefix string) error {

	list := dataStorage.GetListWithoutCollections()
	for _, it := range list {
		it.DeleteTransitionalChangesFromStore(prefix)
	}

	bytes := dataStorage.DBTx.Get("dataStorage:transitionsCollectionsKeys:" + prefix)
	if bytes == nil {
		return errors.New("transitionsCollectionsKeys is null")
	}

	transitionCollectionsKeys := &dataStorageTransitionCollectionsKeys{}
	if err := msgpack.Unmarshal(bytes, transitionCollectionsKeys); err != nil {
		return err
	}

	for _, key := range transitionCollectionsKeys.Accounts {
		accs, err := dataStorage.AccsCollection.GetMap(key)
		if err != nil {
			return err
		}
		accs.DeleteTransitionalChangesFromStore(prefix)
	}

	for _, key := range transitionCollectionsKeys.AssetsFeeLiquidities {
		maxHeap, err := dataStorage.AstsFeeLiquidityCollection.GetMaxHeap(key)
		if err != nil {
			return err
		}
		maxHeap.HashMap.DeleteTransitionalChangesFromStore(prefix)
		maxHeap.DictMap.DeleteTransitionalChangesFromStore(prefix)
	}

	dataStorage.DBTx.Delete("dataStorage:transitionsCollectionsKeys:" + prefix)
	return nil
}
