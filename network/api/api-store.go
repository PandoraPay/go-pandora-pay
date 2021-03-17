package api

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

type BlockWithTxs struct {
	Blk *block.Block
	Txs []helpers.ByteString
}

func (api *API) loadBlockCompleteFromHash(hash []byte) (blkComplete *block_complete.BlockComplete) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		blkComplete = api.loadBlockComplete(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadBlockCompleteFromHeight(blockHeight uint64) (blkComplete *block_complete.BlockComplete) {
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

func (api *API) loadAccountFromPublicKeyHash(publicKeyHash []byte) (acc *account.Account) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		accs := accounts.NewAccounts(boltTx)
		acc = accs.GetAccount(publicKeyHash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadTokenFromPublicKeyHash(publicKeyHash []byte) (tok *token.Token) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		toks := tokens.NewTokens(boltTx)
		tok = toks.GetToken(publicKeyHash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (api *API) loadBlockComplete(bucket *bolt.Bucket, hash []byte) *block_complete.BlockComplete {

	blk := api.chain.LoadBlock(bucket, hash)
	if blk == nil {
		return nil
	}

	txHashes := [][]byte{}
	data := bucket.Get([]byte("blockTxs" + strconv.FormatUint(blk.Height, 10)))
	err := json.Unmarshal(data, &txHashes)
	if err != nil {
		panic(err)
	}

	txs := make([]*transaction.Transaction, len(txHashes))
	for i, txHash := range txHashes {
		data = bucket.Get(append([]byte("tx"), txHash...))
		tx := &transaction.Transaction{}
		tx.Deserialize(helpers.NewBufferReader(data), false)
		txs[i] = tx
	}

	return &block_complete.BlockComplete{
		Block: blk,
		Txs:   txs,
	}
}

func (api *API) loadBlockWithTxHashes(bucket *bolt.Bucket, hash []byte) *BlockWithTxs {
	blk := api.chain.LoadBlock(bucket, hash)
	if blk == nil {
		return nil
	}

	txHashes := [][]byte{}
	data := bucket.Get([]byte("blockTxs" + strconv.FormatUint(blk.Height, 10)))
	err := json.Unmarshal(data, &txHashes)
	if err != nil {
		panic(err)
	}

	txs := make([]helpers.ByteString, len(txHashes))
	for i, txHash := range txHashes {
		txs[i] = txHash
	}

	return &BlockWithTxs{
		Blk: blk,
		Txs: txs,
	}
}

func (api *API) loadTx(bucket *bolt.Bucket, hash []byte) *transaction.Transaction {
	data := bucket.Get(append([]byte("tx"), hash...))
	tx := new(transaction.Transaction)
	tx.Deserialize(helpers.NewBufferReader(data), false)
	return tx
}
