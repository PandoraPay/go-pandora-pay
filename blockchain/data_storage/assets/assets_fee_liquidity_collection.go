package assets

import (
	"errors"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/config/config_coins"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/min_heap"
	"pandora-pay/store/store_db/store_db_interface"
)

//internal
//RED BLACK TREE should be better than MinHeap
type AssetsFeeLiquidityCollection struct {
	tx                store_db_interface.StoreDBTransactionInterface
	liquidityMinHeaps map[string]*min_heap.MinHeapStoreHashMap
	listMaps          []*hash_map.HashMap
}

func (collection *AssetsFeeLiquidityCollection) GetAllMinHeaps() map[string]*min_heap.MinHeapStoreHashMap {
	return collection.liquidityMinHeaps
}

func (collection *AssetsFeeLiquidityCollection) GetAllHashmaps() []*hash_map.HashMap {
	return collection.listMaps
}

func (collection *AssetsFeeLiquidityCollection) GetMinHeap(assetId []byte) (*min_heap.MinHeapStoreHashMap, error) {

	if len(assetId) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Asset was not found")
	}

	if minheap := collection.liquidityMinHeaps[string(assetId)]; minheap != nil {
		return minheap, nil
	}

	minheap := min_heap.NewMinHeapStoreHashMap(collection.tx, string(assetId))
	collection.listMaps = append(collection.listMaps, minheap.HashMap, minheap.DictMap)

	collection.liquidityMinHeaps[string(assetId)] = minheap
	return minheap, nil
}

func (collection *AssetsFeeLiquidityCollection) UpdateLiquidity(publicKey []byte, score uint64, assetId []byte, status asset_fee_liquidity.UpdateLiquidityStatus) error {

	minheap, err := collection.GetMinHeap(assetId)
	if err != nil {
		return err
	}

	switch status {
	case asset_fee_liquidity.UPDATE_LIQUIDITY_OVERWRITTEN:
		if err = minheap.DeleteByKey(publicKey); err != nil {
			return err
		}
		return minheap.Insert(score, publicKey)
	case asset_fee_liquidity.UPDATE_LIQUIDITY_INSERTED:
		return minheap.Insert(score, publicKey)
	case asset_fee_liquidity.UPDATE_LIQUIDITY_DELETED:
		return minheap.DeleteByKey(publicKey)
	default:
		return errors.New("Invalid status")
	}

}

func NewFeeLiquidityCollection(tx store_db_interface.StoreDBTransactionInterface) *AssetsFeeLiquidityCollection {
	return &AssetsFeeLiquidityCollection{
		tx,
		make(map[string]*min_heap.MinHeapStoreHashMap),
		make([]*hash_map.HashMap, 0),
	}
}
