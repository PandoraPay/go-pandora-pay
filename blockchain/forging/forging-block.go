package forging

import (
	"encoding/binary"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"sync/atomic"
	"time"
)

//inside a thread
func forge(threads, threadIndex int) {

	buf := make([]byte, binary.MaxVarintLen64)

	ForgingW.RLock()
	defer ForgingW.RUnlock()
	defer wg.Done()

	serialized := Forging.BlkComplete.Block.SerializeBlock(false, false, false, false, false)
	timestamp := Forging.BlkComplete.Block.Timestamp + 1

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
				serialized = append(serialized, address.delegatedPublicKey[:]...)
				kernelHash := crypto.SHA3Hash(serialized)

				if difficulty.CheckKernelHashBig(kernelHash, Forging.target) {

					Forging.foundSolution(address, timestamp)

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
