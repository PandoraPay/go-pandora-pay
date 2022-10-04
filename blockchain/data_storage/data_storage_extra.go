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
	ConditionalPayments  []uint64
}

func (dataStorage *DataStorage) GetListWithoutCollections() (list []hash_map.HashMapInterface) {
	return []hash_map.HashMapInterface{
		dataStorage.Regs.HashMap,
		dataStorage.PlainAccs.HashMap,
		dataStorage.PendingStakes.HashMap,
		dataStorage.Asts.HashMap,
	}
}

func (dataStorage *DataStorage) GetList(computeChangesSize bool) (list []hash_map.HashMapInterface) {

	list = []hash_map.HashMapInterface{
		dataStorage.Regs.HashMap,
		dataStorage.PlainAccs.HashMap,
		dataStorage.PendingStakes.HashMap,
		dataStorage.Asts.HashMap,
	}

	list = append(list, dataStorage.AccsCollection.GetAllHashmaps()...)

	if !computeChangesSize {
		list = append(list, dataStorage.AstsFeeLiquidityCollection.GetAllHashmaps()...)
		list = append(list, dataStorage.ConditionalPaymentsCollection.GetAllHashmaps()...)
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
	list := dataStorage.GetList(false)
	for _, it := range list {
		it.SetTx(dbTx)
	}
	dataStorage.AccsCollection.SetTx(dbTx)
	dataStorage.AstsFeeLiquidityCollection.SetTx(dbTx)
	dataStorage.ConditionalPaymentsCollection.SetTx(dbTx)
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
		make([]uint64, 0),
	}

	accounts := dataStorage.AccsCollection.GetAllMaps()
	var hasData, hasData2 bool
	for key := range accounts {
		if hasData, err = accounts[key].WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData {
			transitionCollectionsKeys.Accounts = append(transitionCollectionsKeys.Accounts, accounts[key].Asset)
		}
	}

	liquidityMaxHeaps := dataStorage.AstsFeeLiquidityCollection.GetAllMaps()
	for key := range liquidityMaxHeaps {
		if hasData, err = liquidityMaxHeaps[key].HashMap.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData2, err = liquidityMaxHeaps[key].DictMap.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData || hasData2 {
			transitionCollectionsKeys.AssetsFeeLiquidities = append(transitionCollectionsKeys.AssetsFeeLiquidities, []byte(key))
		}
	}

	conditionalPayments := dataStorage.ConditionalPaymentsCollection.GetAllMaps()
	for key := range conditionalPayments {
		if hasData, err = conditionalPayments[key].WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
		if hasData {
			transitionCollectionsKeys.ConditionalPayments = append(transitionCollectionsKeys.ConditionalPayments, conditionalPayments[key].BlockHeight)
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
		hashmap, err := dataStorage.AccsCollection.GetMap(key)
		if err != nil {
			return err
		}
		if err = hashmap.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
	}

	for _, key := range transitionCollectionsKeys.AssetsFeeLiquidities {
		maxHeap, err := dataStorage.AstsFeeLiquidityCollection.GetMaxHeap(key)
		if err != nil {
			return err
		}
		if err = maxHeap.HashMap.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
		if err = maxHeap.DictMap.ReadTransitionalChangesFromStore(prefix); err != nil {
			return err
		}
	}

	for _, key := range transitionCollectionsKeys.ConditionalPayments {
		hashmap, err := dataStorage.ConditionalPaymentsCollection.GetMap(key)
		if err != nil {
			return err
		}
		if err = hashmap.ReadTransitionalChangesFromStore(prefix); err != nil {
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
		hashmap, err := dataStorage.AccsCollection.GetMap(key)
		if err != nil {
			return err
		}
		hashmap.DeleteTransitionalChangesFromStore(prefix)
	}

	for _, key := range transitionCollectionsKeys.AssetsFeeLiquidities {
		maxHeap, err := dataStorage.AstsFeeLiquidityCollection.GetMaxHeap(key)
		if err != nil {
			return err
		}
		maxHeap.HashMap.DeleteTransitionalChangesFromStore(prefix)
		maxHeap.DictMap.DeleteTransitionalChangesFromStore(prefix)
	}

	for _, key := range transitionCollectionsKeys.ConditionalPayments {
		hashmap, err := dataStorage.ConditionalPaymentsCollection.GetMap(key)
		if err != nil {
			return err
		}
		hashmap.DeleteTransitionalChangesFromStore(prefix)
	}

	dataStorage.DBTx.Delete("dataStorage:transitionsCollectionsKeys:" + prefix)
	return nil
}
