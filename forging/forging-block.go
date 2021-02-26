package forging

import (
	"encoding/binary"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"sync"
	"sync/atomic"
	"time"
)

func createNextBlockComplete(height uint64) (blkComplete *block.BlockComplete, err error) {

	var blk *block.Block
	if height == 0 {
		blk, err = genesis.CreateNewGenesisBlock()
		if err != nil {
			return
		}
	} else {

		var blockHeader = block.BlockHeader{
			Version: 0,
			Height:  height,
		}

		blk = &block.Block{
			BlockHeader:    blockHeader,
			MerkleHash:     crypto.SHA3Hash([]byte{}),
			PrevHash:       blockchain.Chain.Hash,
			PrevKernelHash: blockchain.Chain.KernelHash,
			Timestamp:      blockchain.Chain.Timestamp,
		}

	}

	blkComplete = &block.BlockComplete{
		Block: blk,
	}

	return

}

//inside a thread
func forge(threads, threadIndex int, wg *sync.WaitGroup) {

	buf := make([]byte, binary.MaxVarintLen64)

	ForgingW.RLock()
	defer ForgingW.RUnlock()
	defer wg.Done()

	serialized := forging.blkComplete.Block.SerializeBlock(false, false, false, false, false)
	timestamp := forging.blkComplete.Block.Timestamp + 1

	for atomic.LoadInt32(&forgingWorking) == 1 {

		if timestamp > uint64(time.Now().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		//forge with my wallets
		for i, address := range ForgingW.addresses {

			if i%threads == threadIndex {

				if atomic.LoadInt32(&forgingWorking) == 0 {
					break
				}

				n := binary.PutUvarint(buf, timestamp)
				serialized = append(serialized, buf[:n]...)
				serialized = append(serialized, address.publicKey[:]...)
				kernelHash := crypto.SHA3Hash(serialized)

				if difficulty.CheckKernelHashBig(kernelHash, blockchain.Chain.Target) {

					forging.foundSolution(address, timestamp)

				} else {
					// for debugging only
					// gui.Log(hex.EncodeToString(kernelHash[:]))
				}

				serialized = serialized[:len(serialized)-n-33]

			}

		}
		timestamp += 1

	}

}
