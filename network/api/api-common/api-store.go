package api_common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/blocks/block-info"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	token_info "pandora-pay/blockchain/tokens/token-info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

type APIStore struct {
	chain *blockchain.Blockchain
}

func (apiStore *APIStore) LoadBlockInfoFromHash(hash []byte) (blkInfo *block_info.BlockInfo, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		blkInfo, err = apiStore.LoadBlockInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockInfoFromHeight(blockHeight uint64) (blkInfo *block_info.BlockInfo, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		blkInfo, err = apiStore.LoadBlockInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockCompleteFromHash(hash []byte) (blkComplete *block_complete.BlockComplete, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		blkComplete, err = apiStore.LoadBlockComplete(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockCompleteFromHeight(blockHeight uint64) (blkComplete *block_complete.BlockComplete, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		blkComplete, err = apiStore.LoadBlockComplete(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockWithTXsFromHash(hash []byte) (blkWithTXs *api_types.APIBlockWithTxs, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		blkWithTXs, err = apiStore.LoadBlockWithTxHashes(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadTxFromHash(hash []byte) (tx *transaction.Transaction, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		tx, err = apiStore.LoadTx(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadTxFromHeight(txHeight uint64) (tx *transaction.Transaction, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err := apiStore.chain.LoadTxHash(reader, txHeight)
		if err != nil {
			return
		}
		tx, err = apiStore.LoadTx(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockWithTXsFromHeight(blockHeight uint64) (blkWithTXs *api_types.APIBlockWithTxs, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err := apiStore.chain.LoadBlockHash(reader, blockHeight)
		if err != nil {
			return
		}
		blkWithTXs, err = apiStore.LoadBlockWithTxHashes(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) LoadAccountFromPublicKeyHash(publicKeyHash []byte) (acc *account.Account, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)
		acc, err = accs.GetAccount(publicKeyHash, chainHeight)
		return
	})
	return
}

func (apiStore *APIStore) LoadTokenFromPublicKeyHash(publicKeyHash []byte) (tok *token.Token, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		toks := tokens.NewTokens(reader)
		tok, err = toks.GetToken(publicKeyHash)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockHash(blockHeight uint64) (hash []byte, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err = apiStore.chain.LoadBlockHash(reader, blockHeight)
		return
	})
	return
}

func (apiStore *APIStore) LoadTxHash(blockHeight uint64) (hash []byte, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err = apiStore.chain.LoadTxHash(reader, blockHeight)
		return
	})
	return
}

func (apiStore *APIStore) LoadBlockComplete(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (out *block_complete.BlockComplete, err error) {

	blk, err := apiStore.chain.LoadBlock(reader, hash)
	if blk == nil || err != nil {
		return
	}

	data := reader.Get("blockTxs" + strconv.FormatUint(blk.Height, 10))
	if data == nil {
		return
	}

	txHashes := [][]byte{}
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
	}

	txs := make([]*transaction.Transaction, len(txHashes))
	for i, txHash := range txHashes {
		data = reader.Get("tx" + string(txHash))
		txs[i] = &transaction.Transaction{}
		if err = txs[i].Deserialize(helpers.NewBufferReader(data)); err != nil {
			return
		}
	}

	return &block_complete.BlockComplete{
		Block: blk,
		Txs:   txs,
	}, nil
}

func (apiStore *APIStore) LoadBlockWithTxHashes(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (out *api_types.APIBlockWithTxs, err error) {
	blk, err := apiStore.chain.LoadBlock(reader, hash)
	if blk == nil || err != nil {
		return
	}

	txHashes := [][]byte{}
	data := reader.Get("blockTxs" + strconv.FormatUint(blk.Height, 10))
	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
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

func (apiStore *APIStore) LoadBlockInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (blkInfo *block_info.BlockInfo, err error) {
	data := reader.Get("blockInfo_ByHash" + string(hash))
	if data == nil {
		return nil, errors.New("BlockInfo was not found")
	}
	blkInfo = &block_info.BlockInfo{}
	err = json.Unmarshal(data, blkInfo)
	return
}

func (apiStore *APIStore) LoadTokenInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (tokInfo *token_info.TokenInfo, err error) {
	data := reader.Get("tokenInfo_ByHash" + string(hash))
	if data == nil {
		return nil, errors.New("TokenInfo was not found")
	}
	tokInfo = &token_info.TokenInfo{}
	err = json.Unmarshal(data, tokInfo)
	return
}

func (apiStore *APIStore) LoadTx(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (tx *transaction.Transaction, err error) {
	data := reader.Get("tx" + string(hash))
	if data == nil {
		return nil, errors.New("Tx not found")
	}

	tx = new(transaction.Transaction)
	err = tx.Deserialize(helpers.NewBufferReader(data))
	return
}

func CreateAPIStore(chain *blockchain.Blockchain) *APIStore {
	return &APIStore{
		chain: chain,
	}
}
