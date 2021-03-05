package forging

import (
	"encoding/binary"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/config/stake"
	"pandora-pay/cryptography"
	"sync/atomic"
	"time"
)

//inside a thread
func forge(forging *Forging, threads, threadIndex int) {

	buf := make([]byte, binary.MaxVarintLen64)

	forging.Wallet.RLock()
	defer forging.Wallet.RUnlock()
	defer forging.wg.Done()

	height := forging.blkComplete.Block.Height
	serialized := forging.blkComplete.Block.serializeBlockForForging()
	n := binary.PutUvarint(buf, forging.blkComplete.Block.Timestamp)

	serialized = serialized[:len(serialized)-n-20]
	timestamp := forging.blkComplete.Block.Timestamp + 1

	for atomic.LoadInt32(&forging.forgingWorking) == 1 {

		if timestamp > uint64(time.Now().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		//forge with my wallets
		for i, address := range forging.Wallet.addresses {

			if i%threads == threadIndex && (address.account != nil || height == 0) {

				var stakingAmount uint64
				if address.account != nil {
					stakingAmount = address.account.GetDelegatedStakeAvailable(height)
				}

				if stakingAmount >= stake.GetRequiredStake(height) {

					if atomic.LoadInt32(&forging.forgingWorking) == 0 {
						break
					}

					n = binary.PutUvarint(buf, timestamp)
					serialized = append(serialized, buf[:n]...)
					serialized = append(serialized, address.publicKeyHash[:]...)
					kernelHash := cryptography.SHA3Hash(serialized)

					if height > 0 {
						kernelHash = cryptography.ComputeKernelHash(kernelHash, stakingAmount)
					}

					if difficulty.CheckKernelHashBig(kernelHash, forging.target) {

						forging.foundSolution(address, timestamp)

					} else {
						// for debugging only
						// gui.Log(hex.EncodeToString(kernelHash[:]))
					}

					serialized = serialized[:len(serialized)-n-20]

				}

			}

		}
		timestamp += 1

	}

}
