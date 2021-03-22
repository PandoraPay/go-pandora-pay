package api_store

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain"
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

type APIStore struct {
	chain *blockchain.Blockchain
}

func (apiStore *APIStore) LoadBlockCompleteFromHash(hash []byte) (blkComplete *block_complete.BlockComplete) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		blkComplete = apiStore.LoadBlockComplete(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadBlockCompleteFromHeight(blockHeight uint64) (blkComplete *block_complete.BlockComplete) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		hash := apiStore.chain.LoadBlockHash(reader, blockHeight)
		blkComplete = apiStore.LoadBlockComplete(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadBlockWithTXsFromHash(hash []byte) (blkWithTXs *BlockWithTxs) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		blkWithTXs = apiStore.LoadBlockWithTxHashes(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadTxFromHash(hash []byte) (tx *transaction.Transaction) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		tx = apiStore.LoadTx(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadBlockWithTXsFromHeight(blockHeight uint64) (blkWithTXs *BlockWithTxs) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		hash := apiStore.chain.LoadBlockHash(reader, blockHeight)
		blkWithTXs = apiStore.LoadBlockWithTxHashes(reader, hash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadAccountFromPublicKeyHash(publicKeyHash []byte) (acc *account.Account) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		accs := accounts.NewAccounts(boltTx)
		acc = accs.GetAccount(publicKeyHash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadTokenFromPublicKeyHash(publicKeyHash []byte) (tok *token.Token) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		toks := tokens.NewTokens(boltTx)
		tok = toks.GetToken(publicKeyHash)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadBlockHash(blockHeight uint64) (hash []byte) {
	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		hash = apiStore.chain.LoadBlockHash(reader, blockHeight)
		return nil
	}); err != nil {
		panic(err)
	}
	return
}

func (apiStore *APIStore) LoadBlockComplete(bucket *bolt.Bucket, hash []byte) *block_complete.BlockComplete {

	blk := apiStore.chain.LoadBlock(bucket, hash)
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

func (apiStore *APIStore) LoadBlockWithTxHashes(bucket *bolt.Bucket, hash []byte) *BlockWithTxs {
	blk := apiStore.chain.LoadBlock(bucket, hash)
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

func (apiStore *APIStore) LoadTx(bucket *bolt.Bucket, hash []byte) *transaction.Transaction {
	data := bucket.Get(append([]byte("tx"), hash...))
	tx := new(transaction.Transaction)
	tx.Deserialize(helpers.NewBufferReader(data), false)
	return tx
}

func CreateAPIStore(chain *blockchain.Blockchain) *APIStore {
	return &APIStore{
		chain: chain,
	}
}
