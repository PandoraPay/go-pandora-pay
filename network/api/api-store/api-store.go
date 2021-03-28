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

func (apiStore *APIStore) LoadBlockCompleteFromHash(hash []byte) (blkComplete *block_complete.BlockComplete, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		if blkComplete, err = apiStore.LoadBlockComplete(reader, hash); err != nil {
			return
		}
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockCompleteFromHeight(blockHeight uint64) (blkComplete *block_complete.BlockComplete, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		if blkComplete, err = apiStore.LoadBlockComplete(reader, hash); err != nil {
			return
		}
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockWithTXsFromHash(hash []byte) (blkWithTXs *BlockWithTxs, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		if blkWithTXs, err = apiStore.LoadBlockWithTxHashes(reader, hash); err != nil {
			return
		}
		return
	})
	return
}

func (apiStore *APIStore) LoadTxFromHash(hash []byte) (tx *transaction.Transaction, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		tx, err = apiStore.LoadTx(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockWithTXsFromHeight(blockHeight uint64) (blkWithTXs *BlockWithTxs, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		if blkWithTXs, err = apiStore.LoadBlockWithTxHashes(reader, hash); err != nil {
			return
		}
		return
	})
	return
}

func (apiStore *APIStore) LoadAccountFromPublicKeyHash(publicKeyHash []byte) (acc *account.Account, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		accs := accounts.NewAccounts(boltTx)
		acc = accs.GetAccount(publicKeyHash)
		return nil
	})
	return
}

func (apiStore *APIStore) LoadTokenFromPublicKeyHash(publicKeyHash []byte) (tok *token.Token, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		toks := tokens.NewTokens(boltTx)
		tok = toks.GetToken(publicKeyHash)
		return nil
	})
	return
}

func (apiStore *APIStore) LoadBlockHash(blockHeight uint64) (hash []byte, err error) {
	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		hash, err = apiStore.chain.LoadBlockHash(reader, blockHeight)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockComplete(bucket *bolt.Bucket, hash []byte) (out *block_complete.BlockComplete, err error) {

	blk, err := apiStore.chain.LoadBlock(bucket, hash)
	if blk == nil || err != nil {
		return
	}

	txHashes := [][]byte{}
	data := bucket.Get([]byte("blockTxs" + strconv.FormatUint(blk.Height, 10)))
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
	}

	txs := make([]*transaction.Transaction, len(txHashes))
	for i, txHash := range txHashes {
		data = bucket.Get(append([]byte("tx"), txHash...))
		tx := &transaction.Transaction{}
		if err = tx.Deserialize(helpers.NewBufferReader(data), false); err != nil {
			return
		}
		txs[i] = tx
	}

	return &block_complete.BlockComplete{
		Block: blk,
		Txs:   txs,
	}, nil
}

func (apiStore *APIStore) LoadBlockWithTxHashes(bucket *bolt.Bucket, hash []byte) (out *BlockWithTxs, err error) {
	blk, err := apiStore.chain.LoadBlock(bucket, hash)
	if blk == nil || err != nil {
		return
	}

	txHashes := [][]byte{}
	data := bucket.Get([]byte("blockTxs" + strconv.FormatUint(blk.Height, 10)))
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
	}

	txs := make([]helpers.ByteString, len(txHashes))
	for i, txHash := range txHashes {
		txs[i] = txHash
	}

	return &BlockWithTxs{
		Blk: blk,
		Txs: txs,
	}, nil
}

func (apiStore *APIStore) LoadTx(bucket *bolt.Bucket, hash []byte) (tx *transaction.Transaction, err error) {
	data := bucket.Get(append([]byte("tx"), hash...))
	tx = new(transaction.Transaction)
	if err = tx.Deserialize(helpers.NewBufferReader(data), false); err != nil {
		return
	}
	return
}

func CreateAPIStore(chain *blockchain.Blockchain) *APIStore {
	return &APIStore{
		chain: chain,
	}
}
