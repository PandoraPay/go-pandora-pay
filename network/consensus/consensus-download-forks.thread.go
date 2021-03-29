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

	if !fork.readyForDownloading.SetToIf(false, true) {
		return false
	}

	fork.Lock()
	defer fork.Unlock()

	chainData := thread.chain.GetChainData()

	var err error

	if fork.start > chainData.Height {
		fork.start = chainData.Height
		fork.current = chainData.Height
	}

	for {

		if fork.start == 0 || ((chainData.Height-fork.start > config.FORK_MAX_UNCLE_ALLOWED) && (chainData.Height-fork.start > chainData.ConsecutiveSelfForged)) {
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

		answer := conn.SendJSONAwaitAnswer([]byte("block-complete"), api_websockets.APIBlockHeight(fork.start-1))
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

		hash := blkComplete.Block.Bloom.Hash
		fork2Data, exists := thread.forks.hashes.LoadOrStore(string(hash), fork)
		if exists { //let's merge
			fork2 := fork2Data.(*Fork)
			if fork2.mergeFork(fork) {
				return false
			}
		}

		chainHash, err := thread.apiStore.LoadBlockHash(fork.start - 1)
		if err == nil {
			if bytes.Equal(hash, chainHash) {
				return true
			}
		}

		//prepend
		fork.blocks = append(fork.blocks, nil)
		copy(fork.blocks[1:], fork.blocks)
		fork.blocks[0] = blkComplete

		fork.start -= 1
	}

	return true

}

func (thread *ConsensusProcessForksThread) downloadRemainingBlocks(fork *Fork) (result bool, moreToDownload bool) {

	fork.Lock()
	defer fork.Unlock()

	var err error

	for i := uint64(0); i < config.FORK_MAX_DOWNLOAD; i++ {

		if fork.current == fork.end {
			break
		}

		if fork.errors > 2 {
			return
		}
		if fork.errors > -10 {
			fork.errors = -10
		}

		conn := fork.getRandomConn()
		if conn == nil {
			return
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

	result = true
	moreToDownload = fork.current < fork.end
	return

}

func (thread *ConsensusProcessForksThread) execute() {

	for {

		fork := thread.forks.getBestFork(thread.forks.forksDownloadMap)
		if fork != nil {

			downloaded := thread.downloadFork(fork)
			willRemove := true

			if downloaded {

				success, more := thread.downloadRemainingBlocks(fork)
				if success {

					if err := thread.chain.AddBlocks(fork.blocks, false); err != nil {
						gui.Error("Invalid Fork", err)
					} else {
						if more {
							fork.Lock()
							fork.blocks = []*block_complete.BlockComplete{}
							fork.readyForDownloading.UnSet()
							fork.errors = 0
							fork.Unlock()
							willRemove = false
						}
					}

				}

			}

			if willRemove {
				thread.forks.forksDownloadMap.Delete(fork.index)
				thread.forks.removeFork(fork)
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
