package api_common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIStore struct {
	chain *blockchain.Blockchain
}

func (apiStore *APIStore) openLoadAssetInfo(hash []byte, height uint64) (astInfo *info.AssetInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(hash) == 0 {
			if hash, err = apiStore.loadAssetHash(reader, height); err != nil {
				return
			}
		}

		astInfo, err = apiStore.loadAssetInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadTxInfo(hash []byte, txHeight uint64) (txInfo *info.TxInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(hash) == 0 {
			if hash, err = apiStore.loadTxHash(reader, txHeight); err != nil {
				return
			}
		}

		txInfo, err = apiStore.loadTxInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadTxPreview(hash []byte, txHeight uint64) (txPreview *info.TxPreview, txInfo *info.TxInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(hash) == 0 {
			if hash, err = apiStore.loadTxHash(reader, txHeight); err != nil {
				return
			}
		}

		if txPreview, err = apiStore.loadTxPreview(reader, hash); err != nil {
			return
		}
		txInfo, err = apiStore.loadTxInfo(reader, hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadBlockInfo(blockHeight uint64, hash []byte) (blkInfo *info.BlockInfo, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(hash) == 0 {
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

func (apiStore *APIStore) openLoadBlockWithTXsFromHash(hash []byte, blockHeight uint64, returnType api_types.APIReturnType) (blkWithTXs *api_types.APIBlockWithTxs, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		if len(hash) == 0 {
			if hash, err = apiStore.chain.LoadBlockHash(reader, blockHeight); err != nil {
				return
			}
		}
		blkWithTXs, err = apiStore.loadBlockWithTxHashes(reader, hash, returnType)
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

func (apiStore *APIStore) OpenLoadAccountFromPublicKey(publicKey []byte) (*api_types.APIAccount, error) {

	apiAcc := &api_types.APIAccount{}

	if errFinal := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		accsCollection := accounts.NewAccountsCollection(reader)
		plainAccs := plain_accounts.NewPlainAccounts(reader)
		regs := registrations.NewRegistrations(reader)

		assetsList, err := accsCollection.GetAccountAssets(publicKey)
		if err != nil {
			return
		}

		apiAcc.Accs = make([]*account.Account, len(assetsList))
		apiAcc.Assets = make([]helpers.HexBytes, len(assetsList))

		for i, assetId := range assetsList {

			apiAcc.Assets[i] = assetId

			var accs *accounts.Accounts
			if accs, err = accsCollection.GetMap(assetId); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccount(publicKey); err != nil {
				return
			}

			apiAcc.Accs[i] = acc
		}

		if apiAcc.PlainAcc, err = plainAccs.GetPlainAccount(publicKey, chainHeight); err != nil {
			return
		}

		if apiAcc.Reg, err = regs.GetRegistration(publicKey); err != nil {
			return
		}

		return
	}); errFinal != nil {
		return nil, errFinal
	}
	return apiAcc, nil
}

func (apiStore *APIStore) openLoadAccountTxsFromPublicKey(publicKey []byte, next uint64) (answer *api_types.APIAccountTxs, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		publicKeyStr := string(publicKey)

		data := reader.Get("addrTxsCount:" + publicKeyStr)
		if data == nil {
			return nil
		}

		var count uint64
		if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
			return
		}

		if next > count {
			next = count
		}

		index := next
		if index < config.API_ACCOUNT_MAX_TXS {
			index = 0
		} else {
			index -= config.API_ACCOUNT_MAX_TXS
		}

		answer = &api_types.APIAccountTxs{
			Count: count,
			Txs:   make([][]byte, next-index),
		}
		for i := index; i < next; i++ {
			hash := reader.Get("addrTx:" + publicKeyStr + ":" + strconv.FormatUint(i, 10))
			if hash == nil {
				return errors.New("Error reading address transaction")
			}
			answer.Txs[next-i-1] = hash
		}

		return
	})
	return
}

func (apiStore *APIStore) openLoadAsset(hash []byte, height uint64) (ast *asset.Asset, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if hash == nil {
			if hash, err = apiStore.loadAssetHash(reader, height); err != nil {
				return
			}
		}

		asts := assets.NewAssets(reader)
		ast, err = asts.GetAsset(hash)
		return
	})
	return
}

func (apiStore *APIStore) openLoadAccountsCountFromAssetId(hash []byte) (output uint64, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		accsCollection := accounts.NewAccountsCollection(reader)

		accs, err := accsCollection.GetMap(hash)
		if err != nil {
			return
		}

		output = accs.Count
		return
	})
	return
}

func (apiStore *APIStore) openLoadAccountsKeysByIndex(indexes []uint64, assetId []byte) (output [][]byte, errFinal error) {

	if len(indexes) > 512*2 {
		return nil, fmt.Errorf("Too many indexes to process: limit %d, found %d", 512*2, len(indexes))
	}

	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		accsCollection := accounts.NewAccountsCollection(reader)

		accs, err := accsCollection.GetMap(assetId)
		if err != nil {
			return
		}

		output = make([][]byte, len(indexes))
		for i := 0; i < len(indexes); i++ {
			if output[i], err = accs.GetKeyByIndex(indexes[i]); err != nil {
				return
			}
		}

		return
	})
	return
}

func (apiStore *APIStore) openLoadAccountsByKeys(publicKeys [][]byte, assetId []byte) (output *api_types.APIAccountsByKeys, errFinal error) {

	if len(publicKeys) > 512*2 {
		return nil, fmt.Errorf("Too many indexes to process: limit %d, found %d", 512*2, len(publicKeys))
	}

	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accsCollection := accounts.NewAccountsCollection(reader)
		regs := registrations.NewRegistrations(reader)

		accs, err := accsCollection.GetMap(assetId)
		if err != nil {
			return
		}

		output = &api_types.APIAccountsByKeys{
			Acc: make([]*account.Account, len(publicKeys)),
			Reg: make([]*registration.Registration, len(publicKeys)),
		}

		for i := 0; i < len(publicKeys); i++ {
			if output.Acc[i], err = accs.GetAccount(publicKeys[i]); err != nil {
				return
			}
			if output.Reg[i], err = regs.GetRegistration(publicKeys[i]); err != nil {
				return
			}
		}

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

	out.Txs = make([][]byte, len(missingTxs))
	for i, txMissingIndex := range missingTxs {
		if txMissingIndex >= 0 && txMissingIndex < len(txHashes) {
			tx := reader.Get("tx:" + string(txHashes[txMissingIndex]))
			if tx == nil {
				return nil, errors.New("Tx was not found")
			}
			out.Txs[i] = tx
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
		data = reader.Get("tx:" + string(txHash))
		txs[i] = &transaction.Transaction{}
		if err = txs[i].Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
	}

	blkComplete := &block_complete.BlockComplete{
		Block: blk,
		Txs:   txs,
	}

	if err = blkComplete.BloomCompleteBySerialized(blkComplete.SerializeManualToBytes()); err != nil {
		return nil, err
	}

	return blkComplete, nil
}

func (apiStore *APIStore) loadBlockWithTxHashes(reader store_db_interface.StoreDBTransactionInterface, hash []byte, returnType api_types.APIReturnType) (*api_types.APIBlockWithTxs, error) {
	blk, err := apiStore.loadBlock(reader, hash)
	if blk == nil || err != nil {
		return nil, err
	}

	out := &api_types.APIBlockWithTxs{}
	data := reader.Get("blockTxs" + strconv.FormatUint(blk.Height, 10))
	if err = json.Unmarshal(data, &out.Txs); err != nil {
		return nil, err
	}

	switch returnType {
	case api_types.APIReturnType_RETURN_JSON:
		var marshal []byte
		if marshal, err = json.Marshal(blk); err != nil {
			return nil, err
		}
		out.Block = &api_types.APIBlockWithTxs_Json{string(marshal)}
	case api_types.APIReturnType_RETURN_SERIALIZED:
		out.Block = &api_types.APIBlockWithTxs_Serialized{blk.SerializeToBytes()}
	}

	return out, nil
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
		return nil, errors.New("TxInfo was not found")
	}
	txInfo := &info.TxInfo{}
	return txInfo, json.Unmarshal(data, txInfo)
}

func (apiStore *APIStore) loadTxPreview(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*info.TxPreview, error) {
	data := reader.Get("txPreview_ByHash" + string(hash))
	if data == nil {
		return nil, errors.New("TxPreview was not found")
	}
	txPreview := &info.TxPreview{}
	return txPreview, json.Unmarshal(data, txPreview)
}

func (apiStore *APIStore) loadAssetInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*info.AssetInfo, error) {

	data := reader.Get("assetInfo_ByHash:" + string(hash))
	if data == nil {
		return nil, errors.New("AssetInfo was not found")
	}
	astInfo := &info.AssetInfo{}
	return astInfo, json.Unmarshal(data, astInfo)
}

func (apiStore *APIStore) loadTx(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*transaction.Transaction, *info.TxInfo, error) {

	hashStr := string(hash)
	var data []byte

	if data = reader.Get("tx:" + hashStr); data == nil {
		return nil, nil, errors.New("Tx not found")
	}

	tx := &transaction.Transaction{}
	if err := tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return nil, nil, err
	}
	if err := tx.BloomExtraVerified(); err != nil {
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

func (apiStore *APIStore) loadAssetHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}
	return reader.Get("assets::list:" + strconv.FormatUint(height, 10)), nil
}

func (apiStore *APIStore) loadTxHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}
	return reader.Get("txHash_ByHeight" + strconv.FormatUint(height, 10)), nil
}

func (chain *APIStore) loadBlock(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*block.Block, error) {
	blockData := reader.Get("block_ByHash" + string(hash))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk := block.CreateEmptyBlock()
	return blk, blk.Deserialize(helpers.NewBufferReader(blockData))
}

func CreateAPIStore(chain *blockchain.Blockchain) *APIStore {
	return &APIStore{
		chain: chain,
	}
}
