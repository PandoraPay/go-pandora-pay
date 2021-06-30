package consensus

import (
	"bytes"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"time"
)

type ConsensusProcessForksThread struct {
	chain    *blockchain.Blockchain
	forks    *Forks
	mempool  *mempool.Mempool
	apiStore *api_common.APIStore
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

	var err error

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

		answer := conn.SendJSONAwaitAnswer([]byte("block-complete"), api_types.APIBlockCompleteRequest{start - 1, nil, api_types.RETURN_SERIALIZED})
		if answer.Err != nil {
			fork.errors += 1
			continue
		}

		blkComplete := block_complete.CreateEmptyBlockComplete()
		if err = blkComplete.Deserialize(helpers.NewBufferReader(answer.Out)); err != nil {
			fork.errors += 1
			continue
		}
		if err = blkComplete.BloomAll(); err != nil {
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

		if err := func() (err error) {

			answer := conn.SendJSONAwaitAnswer([]byte("block"), &api_types.APIBlockRequest{fork.Current, nil, api_types.RETURN_SERIALIZED})
			if answer.Err != nil {
				return answer.Err
			}

			blkWithTx := &api_types.APIBlockWithTxs{}
			if err = json.Unmarshal(answer.Out, &blkWithTx); err != nil {
				return
			}

			txsFound := 0
			txs := make([]*transaction.Transaction, len(blkWithTx.Txs))
			for i := range txs {
				if tx := thread.mempool.Txs.Exists(string(blkWithTx.Txs[i])); tx != nil {
					txs[i] = tx
					txsFound++
					continue
				}
			}

			if txsFound < len(txs) {

				serializedTxs := make([][]byte, len(txs))

				_ = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

					for i := range txs {
						if txs[i] == nil {
							serialized := reader.Get("tx" + string(blkWithTx.Txs[i]))
							if serialized != nil {
								serializedTxs[i] = helpers.CloneBytes(serialized)
							}
						}
					}
					return
				})

				for i, serializedTx := range serializedTxs {
					if serializedTx != nil {
						tx := &transaction.Transaction{}
						if err = tx.Deserialize(helpers.NewBufferReader(serializedTx)); err != nil {
							return
						}
						if err = tx.BloomExtraVerified(); err != nil {
							return
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

			answer = conn.SendJSONAwaitAnswer([]byte("block-miss-txs"), &api_types.APIBlockCompleteMissingTxsRequest{blkWithTx.Block.Bloom.Hash, missingTxs})
			if answer.Err != nil {
				return answer.Err
			}
			blkCompleteMissingTxs := &api_types.APIBlockCompleteMissingTxs{}

			if err = json.Unmarshal(answer.Out, blkCompleteMissingTxs); err != nil {
				return
			}
			if len(blkCompleteMissingTxs.Txs) != len(missingTxs) {
				return errors.New("blkCompleteMissingTxs.Txs length is not matching")
			}

			for _, missingTx := range blkCompleteMissingTxs.Txs {
				if missingTx == nil {
					return errors.New("blkCompleteMissingTxs.Tx is null")
				}
			}

			for i, missingTx := range missingTxs {
				tx := &transaction.Transaction{}
				if err = tx.Deserialize(helpers.NewBufferReader(blkCompleteMissingTxs.Txs[i])); err != nil {
					return
				}
				txs[missingTx] = tx
			}

			blkComplete.Txs = txs
			if err = blkComplete.BloomAll(); err != nil {
				return
			}

			fork.Blocks = append(fork.Blocks, blkComplete)
			fork.Current += 1

			return
		}(); err != nil {
			fork.errors += 1
			continue
		}

	}

	return len(fork.Blocks) > 0

}

func (thread *ConsensusProcessForksThread) execute() {

	for {

		fork := thread.forks.getBestFork()
		if fork != nil {

			willRemove := true

			gui.GUI.Log("Status. Downloading fork", fork.Hash)
			if thread.downloadFork(fork) {

				gui.GUI.Log("Status. DownloadingRemainingBlocks fork")

				globals.MainEvents.BroadcastEvent("consensus/update", fork)

				if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
					if thread.downloadRemainingBlocks(fork) {

						gui.GUI.Log("Status. AddBlocks fork")

						if err := thread.chain.AddBlocks(fork.Blocks, false); err != nil {
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
