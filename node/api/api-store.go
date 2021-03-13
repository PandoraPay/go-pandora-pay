package api

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

type BlockWithTxs struct {
	Blk *block.Block
	Txs []helpers.ByteString
}

func (api *API) loadBlockCompleteFromHash(hash []byte) (blkComplete *block.BlockComplete) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		blkComplete = api.loadBlockComplete(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadBlockCompleteFromHeight(blockHeight uint64) (blkComplete *block.BlockComplete) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		hash := api.chain.LoadBlockHash(reader, blockHeight)
		blkComplete = api.loadBlockComplete(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadBlockWithTXsFromHash(hash []byte) (blkWithTXs *BlockWithTxs) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		blkWithTXs = api.loadBlockWithTxHashes(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadTxFromHash(hash []byte) (tx *transaction.Transaction) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		tx = api.loadTx(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadBlockWithTXsFromHeight(blockHeight uint64) (blkWithTXs *BlockWithTxs) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		hash := api.chain.LoadBlockHash(reader, blockHeight)
		blkWithTXs = api.loadBlockWithTxHashes(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadBlockComplete(bucket *bolt.Bucket, hash []byte) *block.BlockComplete {

	blk := api.chain.LoadBlock(bucket, hash)
	if blk == nil {
		return nil
	}

	txHashes := make([][]byte, 0)
	data := bucket.Get([]byte("blockTxs" + strconv.FormatUint(blk.Height, 10)))
	err := json.Unmarshal(data, &txHashes)
	if err != nil {
		panic(err)
	}

	txs := make([]*transaction.Transaction, 0)
	for _, txHash := range txHashes {
		data = bucket.Get(append([]byte("tx"), txHash...))
		tx := &transaction.Transaction{}
		tx.Deserialize(helpers.NewBufferReader(data))
		txs = append(txs, tx)
	}

	return &block.BlockComplete{
		Block: blk,
		Txs:   txs,
	}
}

func (api *API) loadBlockWithTxHashes(bucket *bolt.Bucket, hash []byte) *BlockWithTxs {
	blk := api.chain.LoadBlock(bucket, hash)
	if blk == nil {
		return nil
	}

	txHashes := make([][]byte, 0)
	data := bucket.Get([]byte("blockTxs" + strconv.FormatUint(blk.Height, 10)))
	err := json.Unmarshal(data, &txHashes)
	if err != nil {
		panic(err)
	}

	txs := make([]helpers.ByteString, 0)
	for _, txHash := range txHashes {
		txs = append(txs, txHash)
	}

	return &BlockWithTxs{
		Blk: blk,
		Txs: txs,
	}
}

func (api *API) loadTx(bucket *bolt.Bucket, hash []byte) *transaction.Transaction {
	data := bucket.Get(append([]byte("tx"), hash...))
	tx := new(transaction.Transaction)
	tx.Deserialize(helpers.NewBufferReader(data))
	return tx
}
