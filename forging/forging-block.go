package forging

import (
	"encoding/binary"
	"encoding/hex"
	"pandora-pay/block"
	"pandora-pay/block/difficulty"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/wallet"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

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
func forge(blkComplete *block.BlockComplete, threads, threadIndex int, wg *sync.WaitGroup) {

	buf := make([]byte, binary.MaxVarintLen64)

	serialized := blkComplete.Block.SerializeBlock(false, false, false, false, false)
	timestamp := blkComplete.Block.Timestamp + 1

	addresses := wallet.GetAddresses()

	for forging {

		if timestamp > uint64(time.Now().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		//forge with my wallets
		for i := 0; i < len(addresses) && forging; i++ {

			if i%threads == threadIndex {

				n := binary.PutUvarint(buf, timestamp)
				serialized = append(serialized, buf[:n]...)
				serialized = append(serialized, addresses[i].PublicKey...)
				kernelHash := crypto.SHA3Hash(serialized)

				if difficulty.CheckKernelHashBig(kernelHash, blockchain.Chain.Target) {

					mutex.Lock()

					copy(blkComplete.Block.Forger[:], addresses[i].PublicKey[:])
					blkComplete.Block.Timestamp = timestamp
					serializationForSigning := blkComplete.Block.SerializeForSigning()
					signature, _ := addresses[i].PrivateKey.Sign(&serializationForSigning)

					copy(blkComplete.Block.Signature[:], signature)

					var array []*block.BlockComplete
					array = append(array, blkComplete)

					result, err := blockchain.Chain.AddBlocks(array)
					if err == nil && result {
						forging = false
					}

					mutex.Unlock()

				} else {
					gui.Log(hex.EncodeToString(kernelHash[:]))
				}

				serialized = serialized[:len(serialized)-n-33]

			}

		}
		timestamp += 1

	}

	wg.Done()
}
