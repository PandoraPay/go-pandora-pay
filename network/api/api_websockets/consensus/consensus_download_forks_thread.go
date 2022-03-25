package consensus

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/txs_validator"
	"time"
)

type ConsensusProcessForksThread struct {
	chain        *blockchain.Blockchain
	txsValidator *txs_validator.TxsValidator
	forks        *Forks
	mempool      *mempool.Mempool
}

func (thread *ConsensusProcessForksThread) downloadBlockHash(conn *connection.AdvancedConnection, fork *Fork, height uint64) ([]byte, error) {
	answer, err := connection.SendJSONAwaitAnswer[api_common.APIBlockHashReply](conn, []byte("block-hash"), &api_common.APIBlockHashRequest{height}, nil, 0)
	if err != nil {
		return nil, err
	}

	if len(answer.Hash) != cryptography.HashSize {
		return nil, errors.New("Hash size is invalid")
	}

	return answer.Hash, nil
}

func (thread *ConsensusProcessForksThread) downloadBlockComplete(conn *connection.AdvancedConnection, fork *Fork, height uint64) (*block_complete.BlockComplete, error) {

	blkWithTx, err := connection.SendJSONAwaitAnswer[api_common.APIBlockReply](conn, []byte("block"), &api_common.APIBlockRequest{height, nil, api_types.RETURN_SERIALIZED}, nil, 0)
	if err != nil {
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

	blkComplete := block_complete.CreateEmptyBlockComplete()
	blkComplete.Block = blkWithTx.Block

	missingTxsCount := 0
	for _, tx := range txs {
		if tx == nil {
			missingTxsCount += 1
		}
	}

	if missingTxsCount > 0 {
		missingTxs := make([]int, missingTxsCount)
		c := 0
		for i, tx := range txs {
			if tx == nil {
				missingTxs[c] = i
				c++
			}
		}

		blkCompleteMissingTxs, err := connection.SendJSONAwaitAnswer[APIBlockCompleteMissingTxsReply](conn, []byte("block-miss-txs"), &APIBlockCompleteMissingTxsRequest{blkWithTx.Block.Bloom.Hash, missingTxs}, nil, 0)
		if err != nil {
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
			txs[missingTx] = tx
		}
	}

	blkComplete.Txs = txs

	if err = thread.txsValidator.ValidateTxs(txs); err != nil {
		return nil, err
	}

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

		if start == 0 || ((chainData.Height-start > config.FORK_MAX_UNCLE_ALLOWED+chainData.ConsecutiveSelfForged) && (chainData.Height-start > chainData.ConsecutiveSelfForged)) {
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

		hash, err := thread.downloadBlockHash(conn, fork, start-1)
		if err != nil {
			fork.errors += 1
			continue
		}

		chainHash, err := thread.chain.OpenLoadBlockHash(start - 1)
		if err == nil && bytes.Equal(hash, chainHash) {
			break
		}

		blkComplete, err := thread.downloadBlockComplete(conn, fork, start-1)
		if err != nil {
			fork.errors += 1
			continue
		}

		if !bytes.Equal(blkComplete.Bloom.Hash, hash) { //it is not the same block
			fork.errors += 1
			continue
		}

		//prepend
		fork.Blocks.PushFront(blkComplete)

		start -= 1
	}

	fork.Current = start + uint64(fork.Blocks.Length)

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

		fork.Blocks.Push(blkComplete)
		fork.Current += 1

	}

	return fork.Blocks.Length > 0

}

func (thread *ConsensusProcessForksThread) execute() {

	for {

		fork := thread.forks.getBestFork()
		if fork != nil {

			willRemove := true

			if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {

				if thread.downloadFork(fork) {

					globals.MainEvents.BroadcastEvent("consensus/update", fork)

					if thread.downloadRemainingBlocks(fork) {

						blocks := make([]*block_complete.BlockComplete, fork.Blocks.Length)
						it := fork.Blocks.Head
						i := 0
						for it != nil {
							blocks[i] = it.Data
							i += 1
							it = it.Next
						}

						if _, err := thread.chain.AddBlocks(blocks, false, advanced_connection_types.UUID_ALL); err != nil {
							if config.DEBUG {
								gui.GUI.Error("Invalid Fork", err)
							}
						} else {
							fork.Lock()
							if fork.Current < fork.End {
								fork.Blocks.Empty()
								fork.errors = 0
								willRemove = false
							}
							fork.Unlock()
						}

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

func newConsensusProcessForksThread(forks *Forks, chain *blockchain.Blockchain, mempool *mempool.Mempool, txsValidator *txs_validator.TxsValidator) *ConsensusProcessForksThread {
	return &ConsensusProcessForksThread{
		chain,
		txsValidator,
		forks,
		mempool,
	}
}
