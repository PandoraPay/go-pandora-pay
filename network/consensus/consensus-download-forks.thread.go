package consensus

import (
	"bytes"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common"
	"time"
)

type ConsensusProcessForksThread struct {
	chain    *blockchain.Blockchain
	forks    *Forks
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

		answer := conn.SendJSONAwaitAnswer([]byte("block-complete"), api_common.APIBlockCompleteRequest{start - 1, nil, api_common.RETURN_SERIALIZED})
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

		chainHash, err := thread.apiStore.LoadBlockHash(start - 1)
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

	var err error

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

		answer := conn.SendJSONAwaitAnswer([]byte("block-complete"), &api_common.APIBlockCompleteRequest{fork.Current, nil, api_common.RETURN_SERIALIZED})

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

			gui.GUI.Log("Status. Downloading fork")
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
				thread.forks.removeFork(fork, true)
			}

		}

		time.Sleep(25 * time.Millisecond)
	}
}

func createConsensusProcessForksThread(forks *Forks, chain *blockchain.Blockchain, apiStore *api_common.APIStore) *ConsensusProcessForksThread {
	return &ConsensusProcessForksThread{
		forks:    forks,
		chain:    chain,
		apiStore: apiStore,
	}
}
