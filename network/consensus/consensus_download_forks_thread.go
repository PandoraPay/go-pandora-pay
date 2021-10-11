package consensus

import (
	"bytes"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store_db/store_db_interface"
	"time"
)

type ConsensusProcessForksThread struct {
	chain    *blockchain.Blockchain
	forks    *Forks
	mempool  *mempool.Mempool
	apiStore *api_common.APIStore
}

func (thread *ConsensusProcessForksThread) downloadBlockComplete(conn *connection.AdvancedConnection, fork *Fork, height uint64) (*block_complete.BlockComplete, error) {

	var err error

	answer := conn.SendJSONAwaitAnswer([]byte("block"), &api_types.APIBlockRequest{height, nil, api_types.RETURN_SERIALIZED}, 0)
	if answer.Err != nil {
		return nil, answer.Err
	}

	blkWithTx := &api_types.APIBlockWithTxs{}
	if err = json.Unmarshal(answer.Out, &blkWithTx); err != nil {
		return nil, err
	}
	blkWithTx.Block = block.CreateEmptyBlock()
	if err = blkWithTx.Block.Deserialize(helpers.NewBufferReader(blkWithTx.BlockSerialized)); err != nil {
		return nil, err
	}

	txsFound := 0
	txs := make([]*transaction.Transaction, len(blkWithTx.Txs))
	for i := range txs {
		if tx := thread.mempool.Txs.Get(string(blkWithTx.Txs[i])); tx != nil {
			txs[i] = tx.Tx
			txsFound++
			continue
		}
	}

	if txsFound < len(txs) {

		serializedTxs := make([][]byte, len(txs))

		_ = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
			for i := range txs {
				if txs[i] == nil {
					serializedTxs[i] = reader.GetClone("tx:" + string(blkWithTx.Txs[i]))
				}
			}
			return
		})

		for i, serializedTx := range serializedTxs {
			if serializedTx != nil {
				tx := &transaction.Transaction{}
				if err = tx.Deserialize(helpers.NewBufferReader(serializedTx)); err != nil {
					return nil, err
				}
				if err = tx.BloomExtraVerified(); err != nil {
					return nil, err
				}
				txs[i] = tx
			}
		}

	}

	blkComplete := block_complete.CreateEmptyBlockComplete()
	blkComplete.Block = blkWithTx.Block

	missingTxsCount := 0
	for _, tx := range txs {
		if tx == nil {
			missingTxsCount += 1
		}
	}

	missingTxs := make([]int, missingTxsCount)
	c := 0
	for i, tx := range txs {
		if tx == nil {
			missingTxs[c] = i
			c++
		}
	}

	answer = conn.SendJSONAwaitAnswer([]byte("block-miss-txs"), &api_types.APIBlockCompleteMissingTxsRequest{blkWithTx.Block.Bloom.Hash, missingTxs}, 0)
	if answer.Err != nil {
		return nil, answer.Err
	}
	blkCompleteMissingTxs := &api_types.APIBlockCompleteMissingTxs{}

	if err = json.Unmarshal(answer.Out, blkCompleteMissingTxs); err != nil {
		return nil, err
	}
	if len(blkCompleteMissingTxs.Txs) != len(missingTxs) {
		return nil, errors.New("blkCompleteMissingTxs.Txs length is not matching")
	}

	for _, missingTx := range blkCompleteMissingTxs.Txs {
		if missingTx == nil {
			return nil, errors.New("blkCompleteMissingTxs.Tx is null")
		}
	}

	for i, missingTx := range missingTxs {
		tx := &transaction.Transaction{}
		if err = tx.Deserialize(helpers.NewBufferReader(blkCompleteMissingTxs.Txs[i])); err != nil {
			return nil, err
		}
		if err = tx.BloomExtraVerified(); err != nil {
			return nil, err
		}
		txs[missingTx] = tx
	}

	blkComplete.Txs = txs
	if err = blkComplete.BloomAll(); err != nil {
		return nil, err
	}

	return blkComplete, nil
}

func (thread *ConsensusProcessForksThread) downloadFork(fork *Fork) bool {

	fork.Lock()
	defer fork.Unlock()

	chainData := thread.chain.GetChainData()
	if fork.BigTotalDifficulty.Cmp(chainData.BigTotalDifficulty) <= 0 {
		return false
	}

	if fork.Initialized {
		return true
	}

	start := fork.End
	if start > chainData.Height {
		start = chainData.Height
	}

	for {

		if start == 0 || ((chainData.Height-start > config.FORK_MAX_UNCLE_ALLOWED) && (chainData.Height-start > chainData.ConsecutiveSelfForged)) {
			break
		}

		if fork.errors > 2 {
			return false
		}

		if fork.errors < -10 {
			fork.errors = -10
		}

		conn := fork.getRandomConn()
		if conn == nil {
			return false
		}

		blkComplete, err := thread.downloadBlockComplete(conn, fork, start-1)
		if err != nil {
			fork.errors += 1
			continue
		}

		chainHash, err := thread.chain.OpenLoadBlockHash(start - 1)
		if err == nil && bytes.Equal(blkComplete.Block.Bloom.Hash, chainHash) {
			break
		}

		//prepend
		fork.Blocks = append(fork.Blocks, nil)
		copy(fork.Blocks[1:], fork.Blocks)
		fork.Blocks[0] = blkComplete

		start -= 1
	}

	fork.Current = start + uint64(len(fork.Blocks))

	fork.Initialized = true

	return true
}

func (thread *ConsensusProcessForksThread) downloadRemainingBlocks(fork *Fork) bool {

	fork.Lock()
	defer fork.Unlock()

	for i := uint64(0); i < config.FORK_MAX_DOWNLOAD; i++ {

		if fork.Current == fork.End {
			break
		}

		if fork.errors > 2 {
			return false
		}
		if fork.errors < -10 {
			fork.errors = -10
		}

		conn := fork.getRandomConn()
		if conn == nil {
			return false
		}

		blkComplete, err := thread.downloadBlockComplete(conn, fork, fork.Current)
		if err != nil {
			fork.errors += 1
			continue
		}

		fork.Blocks = append(fork.Blocks, blkComplete)
		fork.Current += 1

	}

	return len(fork.Blocks) > 0

}

func (thread *ConsensusProcessForksThread) execute() {

	for {

		fork := thread.forks.getBestFork()
		if fork != nil {

			willRemove := true

			gui.GUI.Log("Status. Downloading fork", fork.Hash)
			if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {

				if thread.downloadFork(fork) {

					gui.GUI.Log("Status. DownloadingRemainingBlocks fork")

					globals.MainEvents.BroadcastEvent("consensus/update", fork)

					if thread.downloadRemainingBlocks(fork) {

						gui.GUI.Log("Status. AddBlocks fork")

						if err := thread.chain.AddBlocks(fork.Blocks, false, advanced_connection_types.UUID_ALL); err != nil {
							gui.GUI.Error("Invalid Fork", err)
						} else {
							fork.Lock()
							if fork.Current < fork.End {
								fork.Blocks = []*block_complete.BlockComplete{}
								fork.errors = 0
								willRemove = false
							}
							fork.Unlock()
						}
						gui.GUI.Log("Status. AddBlocks DONE fork")

					}
				}

			} else {
				globals.MainEvents.BroadcastEvent("consensus/update", fork)
				gui.GUI.Log("Status. AddBlocks fork - Simulating block")

				newChainData := &blockchain.BlockchainData{
					Height:             fork.End,
					Hash:               fork.Hash,
					PrevHash:           fork.PrevHash,
					BigTotalDifficulty: fork.BigTotalDifficulty,
				}

				thread.chain.ChainData.Store(newChainData)
				thread.mempool.UpdateWork(fork.Hash, fork.End)
			}

			if willRemove {
				thread.forks.removeFork(fork)
			}

		}

		time.Sleep(25 * time.Millisecond)
	}
}

func createConsensusProcessForksThread(forks *Forks, chain *blockchain.Blockchain, mempool *mempool.Mempool, apiStore *api_common.APIStore) *ConsensusProcessForksThread {
	return &ConsensusProcessForksThread{
		chain,
		forks,
		mempool,
		apiStore,
	}
}
