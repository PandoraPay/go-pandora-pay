package api_common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

type APIStore struct {
	chain *blockchain.Blockchain
}

func (apiStore *APIStore) openLoadTokenInfo(hash []byte) (tokInfo *info.TokenInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		tokInfo, err = apiStore.loadTokenInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadTxInfo(hash []byte, txHeight uint64) (txInfo *info.TxInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if hash == nil {
			if hash, err = apiStore.loadTxHash(reader, txHeight); err != nil {
				return
			}
		}

		txInfo, err = apiStore.loadTxInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockInfo(blockHeight uint64, hash []byte) (blkInfo *info.BlockInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if hash == nil {
			if hash, err = apiStore.chain.LoadBlockHash(reader, blockHeight); err != nil {
				return
			}
		}

		blkInfo, err = apiStore.loadBlockInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockCompleteMissingTxsFromHash(hash []byte, missingTxs []int) (blockCompleteMissingTxs *api_types.APIBlockCompleteMissingTxs, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		blockCompleteMissingTxs, err = apiStore.loadBlockCompleteMissingTxs(reader, hash, missingTxs)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockCompleteFromHash(hash []byte) (blkComplete *block_complete.BlockComplete, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		blkComplete, err = apiStore.loadBlockComplete(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockCompleteFromHeight(blockHeight uint64) (blkComplete *block_complete.BlockComplete, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		blkComplete, err = apiStore.loadBlockComplete(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockWithTXsFromHash(hash []byte) (blkWithTXs *api_types.APIBlockWithTxs, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		blkWithTXs, err = apiStore.loadBlockWithTxHashes(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadTx(hash []byte, txHeight uint64) (tx *transaction.Transaction, txInfo *info.TxInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if hash == nil {
			if hash, err = apiStore.loadTxHash(reader, txHeight); err != nil {
				return
			}
		}

		tx, txInfo, err = apiStore.loadTx(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockWithTXsFromHeight(blockHeight uint64) (blkWithTXs *api_types.APIBlockWithTxs, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		blkWithTXs, err = apiStore.loadBlockWithTxHashes(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadAccountFromPublicKeyHash(publicKeyHash []byte) (acc *account.Account, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)
		acc, err = accs.GetAccount(publicKeyHash, chainHeight)
		return
	})
	return
}

func (apiStore *APIStore) openLoadAccountTxsFromPublicKeyHash(publicKeyHash []byte, next uint64) (answer *api_types.APIAccountTxs, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		data := reader.Get("addrTxsCount:" + string(publicKeyHash))
		if data == nil {
			return nil
		}

		var count uint64
		if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
			return
		}
		if next == 0 {
			next = count
		} else if next > count {
			return errors.New("Index exceeding max txs")
		}

		index := next
		if next >= config.API_ACCOUNT_MAX_TXS {
			index -= config.API_ACCOUNT_MAX_TXS
		}

		answer = &api_types.APIAccountTxs{
			Count: count,
			Txs:   make([]helpers.HexBytes, next-index),
		}
		for i := index; i < next; i++ {
			hash := reader.Get("addrTx:" + strconv.FormatUint(i, 10))
			if hash == nil {
				return errors.New("Error reading address transaction")
			}
			answer.Txs[next-i-1] = hash
		}

		return
	})
	return
}

func (apiStore *APIStore) openLoadTokenFromPublicKeyHash(publicKeyHash []byte) (tok *token.Token, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		toks := tokens.NewTokens(reader)
		tok, err = toks.GetToken(publicKeyHash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadTxHash(blockHeight uint64) (hash []byte, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err = apiStore.loadTxHash(reader, blockHeight)
		return
	})
	return
}

func (apiStore *APIStore) loadBlockCompleteMissingTxs(reader store_db_interface.StoreDBTransactionInterface, hash []byte, missingTxs []int) (out *api_types.APIBlockCompleteMissingTxs, err error) {

	heightStr := reader.Get("blockHeight_ByHash" + string(hash))
	if heightStr == nil {
		return nil, errors.New("Block was not found by hash")
	}

	var height uint64
	if height, err = strconv.ParseUint(string(heightStr), 10, 64); err != nil {
		return
	}

	out = &api_types.APIBlockCompleteMissingTxs{}
	data := reader.Get("blockTxs" + strconv.FormatUint(height, 10))
	if data == nil {
		return nil, nil
	}

	txHashes := [][]byte{}
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return nil, err
	}

	out.Txs = make([]helpers.HexBytes, len(missingTxs))
	for i, txMissingIndex := range missingTxs {
		if txMissingIndex >= 0 && txMissingIndex < len(txHashes) {
			out.Txs[i] = txHashes[txMissingIndex]
		}
	}

	return
}

func (apiStore *APIStore) loadBlockComplete(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*block_complete.BlockComplete, error) {

	blk, err := apiStore.loadBlock(reader, hash)
	if blk == nil || err != nil {
		return nil, err
	}

	data := reader.Get("blockTxs" + strconv.FormatUint(blk.Height, 10))
	if data == nil {
		return nil, nil
	}

	txHashes := [][]byte{}
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return nil, err
	}

	txs := make([]*transaction.Transaction, len(txHashes))
	for i, txHash := range txHashes {
		data = reader.Get("tx" + string(txHash))
		txs[i] = &transaction.Transaction{}
		if err = txs[i].Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
	}

	return &block_complete.BlockComplete{
		Block: blk,
		Txs:   txs,
	}, nil
}

func (apiStore *APIStore) loadBlockWithTxHashes(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*api_types.APIBlockWithTxs, error) {
	blk, err := apiStore.loadBlock(reader, hash)
	if blk == nil || err != nil {
		return nil, err
	}

	txHashes := [][]byte{}
	data := reader.Get("blockTxs" + strconv.FormatUint(blk.Height, 10))
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return nil, err
	}

	txs := make([]helpers.HexBytes, len(txHashes))
	for i, txHash := range txHashes {
		txs[i] = txHash
	}

	return &api_types.APIBlockWithTxs{
		Block: blk,
		Txs:   txs,
	}, nil
}

func (apiStore *APIStore) loadBlockInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*info.BlockInfo, error) {
	data := reader.Get("blockInfo_ByHash" + string(hash))
	if data == nil {
		return nil, errors.New("BlockInfo was not found")
	}
	blkInfo := &info.BlockInfo{}
	return blkInfo, json.Unmarshal(data, blkInfo)
}

func (apiStore *APIStore) loadTxInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*info.TxInfo, error) {
	data := reader.Get("txInfo_ByHash" + string(hash))
	if data == nil {
		return nil, errors.New("BlockInfo was not found")
	}
	txInfo := &info.TxInfo{}
	return txInfo, json.Unmarshal(data, txInfo)
}

func (apiStore *APIStore) loadTokenInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*info.TokenInfo, error) {
	if len(hash) == 0 {
		hash = config.NATIVE_TOKEN_FULL
	}
	data := reader.Get("tokenInfo_ByHash" + string(hash))
	if data == nil {
		return nil, errors.New("TokenInfo was not found")
	}
	tokInfo := &info.TokenInfo{}
	return tokInfo, json.Unmarshal(data, tokInfo)
}

func (apiStore *APIStore) loadTx(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*transaction.Transaction, *info.TxInfo, error) {

	hashStr := string(hash)
	var data []byte

	if data = reader.Get("tx" + hashStr); data == nil {
		return nil, nil, errors.New("Tx not found")
	}

	tx := &transaction.Transaction{}
	if err := tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return nil, nil, err
	}

	var txInfo *info.TxInfo
	if config.SEED_WALLET_NODES_INFO {
		if data = reader.Get("txInfo_ByHash" + hashStr); data == nil {
			return nil, nil, errors.New("TxInfo was not found")
		}
		txInfo = &info.TxInfo{}
		if err := json.Unmarshal(data, txInfo); err != nil {
			return nil, nil, err
		}
	}

	return tx, txInfo, nil
}

func (apiStore *APIStore) loadTxHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}

	hash := reader.Get("txHash_ByHeight" + strconv.FormatUint(height, 10))
	return hash, nil
}

func (chain *APIStore) loadBlock(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*block.Block, error) {
	blockData := reader.Get("block_ByHash" + string(hash))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk := &block.Block{BlockHeader: &block.BlockHeader{}}
	return blk, blk.Deserialize(helpers.NewBufferReader(blockData))
}

func CreateAPIStore(chain *blockchain.Blockchain) *APIStore {
	return &APIStore{
		chain: chain,
	}
}
