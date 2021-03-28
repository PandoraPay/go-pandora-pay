package consensus

import (
	"bytes"
	"math/big"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config"
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
	prevHash := fork.prevHash

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
		if fork.errors >= -10 {
			fork.errors = -10
		}

		fork2Data, exists := thread.forks.hashes.LoadOrStore(string(prevHash), fork)
		if exists { //let's merge
			fork2 := fork2Data.(*Fork)
			if fork2.mergeFork(fork) {
				return false
			}
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

		chainHash, err := thread.apiStore.LoadBlockHash(fork.start - 1)
		if err == nil {
			if bytes.Equal(prevHash, chainHash) {
				return true
			}
		}

		fork.blocks = append([]*block_complete.BlockComplete{blkComplete}, fork.blocks...)
		fork.start -= 1
	}

	return true

}

func (thread *ConsensusProcessForksThread) downloadRemainingBlocks(fork *Fork) bool {

	fork.Lock()
	defer fork.Unlock()

	chainData := thread.chain.GetChainData()
	var err error

	for i := uint64(0); i < config.FORK_MAX_DOWNLOAD; i++ {

		if fork.current == fork.end {
			break
		}

		if fork.errors > 2 {
			return false
		}
		if fork.errors >= -10 {
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

	}

	return true

}

func (thread *ConsensusProcessForksThread) execute() {

	for {

		fork := thread.forks.getBestFork(thread.forks.forksDownloadMap)
		if fork != nil {

			downloaded := thread.downloadFork(fork)
			thread.forks.forksDownloadMap.Delete(fork.index)

			if downloaded {

				thread.downloadRemainingBlocks(fork)

				if err := thread.chain.AddBlocks(fork.blocks, false); err != nil {

				}

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
