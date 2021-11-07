package assets

import (
	"errors"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/config/config_coins"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/min_max_heap"
	"pandora-pay/store/store_db/store_db_interface"
)

//internal
//RED BLACK TREE should be better than Heap
type AssetsFeeLiquidityCollection struct {
	tx                store_db_interface.StoreDBTransactionInterface
	liquidityMaxHeaps map[string]*min_max_heap.HeapStoreHashMap
	listMaps          []*hash_map.HashMap
}

func (collection *AssetsFeeLiquidityCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
}

func (collection *AssetsFeeLiquidityCollection) GetAllHashmaps() []*hash_map.HashMap {
	return collection.listMaps
}

func (collection *AssetsFeeLiquidityCollection) GetMaxHeap(assetId []byte) (*min_max_heap.HeapStoreHashMap, error) {

	if len(assetId) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Asset was not found")
	}

	if maxheap := collection.liquidityMaxHeaps[string(assetId)]; maxheap != nil {
		return maxheap, nil
	}

	maxheap := min_max_heap.NewMaxHeapStoreHashMap(collection.tx, string(assetId))
	collection.listMaps = append(collection.listMaps, maxheap.HashMap, maxheap.DictMap)

	collection.liquidityMaxHeaps[string(assetId)] = maxheap
	return maxheap, nil
}

func (collection *AssetsFeeLiquidityCollection) UpdateLiquidity(publicKey []byte, score uint64, assetId []byte, status asset_fee_liquidity.UpdateLiquidityStatus) error {

	maxheap, err := collection.GetMaxHeap(assetId)
	if err != nil {
		return err
	}

	switch status {
	case asset_fee_liquidity.UPDATE_LIQUIDITY_OVERWRITTEN:
		if err = maxheap.DeleteByKey(publicKey); err != nil {
			return err
		}
		return maxheap.Insert(score, publicKey)
	case asset_fee_liquidity.UPDATE_LIQUIDITY_INSERTED:
		return maxheap.Insert(score, publicKey)
	case asset_fee_liquidity.UPDATE_LIQUIDITY_DELETED:
		return maxheap.DeleteByKey(publicKey)
	default:
		return errors.New("Invalid status")
	}

}

func (collection *AssetsFeeLiquidityCollection) GetTopLiquidity(assetId []byte) ([]byte, error) {
	maxheap, err := collection.GetMaxHeap(assetId)
	if err != nil {
		return nil, err
	}

	top, err := maxheap.GetTop()
	if top == nil || err != nil {
		return nil, err
	}

	return top.Key, nil
}

func NewAssetsFeeLiquidityCollection(tx store_db_interface.StoreDBTransactionInterface) *AssetsFeeLiquidityCollection {
	return &AssetsFeeLiquidityCollection{
		tx,
		make(map[string]*min_max_heap.HeapStoreHashMap),
		make([]*hash_map.HashMap, 0),
	}
}
