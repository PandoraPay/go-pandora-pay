package consensus

import (
	"bytes"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config"
	"pandora-pay/gui"
	api_store "pandora-pay/network/api/api-store"
	api_websockets "pandora-pay/network/api/api-websockets"
	"time"
)

type ConsensusProcessForksThread struct {
	chain    *blockchain.Blockchain
	forks    *Forks
	apiStore *api_store.APIStore
}

func (thread *ConsensusProcessForksThread) downloadFork(fork *Fork) bool {

	fork.Lock()
	defer fork.Unlock()

	chainData := thread.chain.GetChainData()
	if fork.bigTotalDifficulty.Cmp(chainData.BigTotalDifficulty) <= 0 {
		return false
	}

	if fork.downloaded {
		return true
	}

	var err error

	start := fork.end
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
		if fork.errors > -10 {
			fork.errors = -10
		}

		conn := fork.getRandomConn()
		if conn == nil {
			return false
		}

		answer := conn.SendJSONAwaitAnswer([]byte("block-complete"), api_websockets.APIBlockHeight(start-1))
		if answer.Err != nil {
			fork.errors += 1
			continue
		}

		blkComplete := block_complete.CreateEmptyBlockComplete()
		if err = blkComplete.Deserialize(answer.Out); err != nil {
			fork.errors += 1
			continue
		}
		if err = blkComplete.BloomAll(); err != nil {
			fork.errors += 1
			continue
		}

		fork.hashes = append(fork.hashes, blkComplete.Block.Bloom.Hash)

		chainHash, err := thread.apiStore.LoadBlockHash(start - 1)
		if err == nil {
			if bytes.Equal(blkComplete.Block.Bloom.Hash, chainHash) {
				break
			}
		}

		//prepend
		fork.blocks = append(fork.blocks, nil)
		copy(fork.blocks[1:], fork.blocks)
		fork.blocks[0] = blkComplete

		start -= 1
	}

	if fork.current == 0 {
		fork.current = start
	}
	fork.downloaded = true

	return true
}

func (thread *ConsensusProcessForksThread) downloadRemainingBlocks(fork *Fork) bool {

	fork.Lock()
	defer fork.Unlock()

	var err error

	for i := uint64(0); i < config.FORK_MAX_DOWNLOAD; i++ {

		if fork.current == fork.end {
			break
		}

		if fork.errors > 2 {
			return false
		}
		if fork.errors > -10 {
			fork.errors = -10
		}

		conn := fork.getRandomConn()
		if conn == nil {
			return false
		}

		answer := conn.SendJSONAwaitAnswer([]byte("block-complete"), api_websockets.APIBlockHeight(fork.current))
		if answer.Err != nil {
			fork.errors += 1
			continue
		}

		blkComplete := block_complete.CreateEmptyBlockComplete()
		if err = blkComplete.Deserialize(answer.Out); err != nil {
			fork.errors += 1
			continue
		}
		if err = blkComplete.BloomAll(); err != nil {
			fork.errors += 1
			continue
		}

		fork.blocks = append(fork.blocks, blkComplete)

		fork.current += 1
	}

	return true

}

func (thread *ConsensusProcessForksThread) execute() {

	for {

		fork := thread.forks.getBestFork()
		if fork != nil {

			willRemove := true

			if thread.downloadFork(fork) {

				if thread.downloadRemainingBlocks(fork) {

					if err := thread.chain.AddBlocks(fork.blocks, false); err != nil {
						gui.Error("Invalid Fork", err)
					} else {
						fork.Lock()
						if fork.current < fork.end {
							fork.blocks = []*block_complete.BlockComplete{}
							fork.errors = 0
							willRemove = false
						}
						fork.Unlock()
					}

				}

			}

			if willRemove {
				thread.forks.removeFork(fork, true)
			}

		}

		time.Sleep(25 * time.Millisecond)
	}
}

func createConsensusProcessForksThread(forks *Forks, chain *blockchain.Blockchain, apiStore *api_store.APIStore) *ConsensusProcessForksThread {
	return &ConsensusProcessForksThread{
		forks:    forks,
		chain:    chain,
		apiStore: apiStore,
	}
}
