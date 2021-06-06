package forging

import (
	"encoding/binary"
	difficulty "pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"sync/atomic"
	"time"
)

type ForgingSolution struct {
	timestamp     uint64
	address       *ForgingWalletAddress
	work          *forging_block_work.ForgingWork
	stakingAmount uint64
}
type ForgingWalletAddressRequired struct {
	publicKeyHash helpers.HexBytes //20 byte
	wallet        *ForgingWalletAddress
	stakingAmount uint64
}

type ForgingWorkerThread struct {
	hashes           uint32
	index            int
	workCn           chan *forging_block_work.ForgingWork // SAFE
	workerSolutionCn chan *ForgingSolution                // SAFE
	walletsCn        chan []*ForgingWalletAddressRequired // SAFE
}

/**
"Staking multiple wallets simultaneously"
*/
func (worker *ForgingWorkerThread) forge() {

	var work *forging_block_work.ForgingWork
	var wallets []*ForgingWalletAddressRequired

	var timestampMs int64
	var timestamp uint64
	var serialized []byte
	var n int
	buf := make([]byte, binary.MaxVarintLen64)

	for {

		timeLimitMs := time.Now().UnixNano()/1000000 + config.NETWORK_TIMESTAMP_DRIFT_MAX_INT*1000
		timeLimit := uint64(timeLimitMs / 1000)

		select {
		case newWork, ok := <-worker.workCn: //or the work was changed meanwhile
			if !ok {
				return
			}
			work = newWork

			writer := helpers.NewBufferWriter()
			work.BlkComplete.Block.SerializeForForging(writer)
			serialized = writer.Bytes()

			n = binary.PutUvarint(buf, work.BlkComplete.Block.Timestamp)

			timestamp = work.BlkComplete.Block.Timestamp + 1
			timestampMs = int64(timestamp) * 1000

		case newWallets := <-worker.walletsCn:
			wallets = newWallets
		default:
		}

		if work == nil || len(wallets) == 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		} else if timestampMs > timeLimitMs {
			time.Sleep(time.Millisecond * time.Duration(timestampMs-timeLimitMs))
			continue
		}

		//forge with my wallets
		diff := int(timeLimit - timestamp)
		if diff > 20 {
			diff = 20
		}
		for i := 0; i <= diff; i++ {
			for _, address := range wallets {

				n2 := binary.PutUvarint(buf, timestamp)

				if n2 != n {
					newSerialized := make([]byte, len(serialized)-n+n2)
					copy(newSerialized, serialized[:-n-cryptography.PublicKeyHashHashSize])
					serialized = newSerialized
					n = n2
				}

				//optimized POS
				copy(serialized[len(serialized)-cryptography.PublicKeyHashHashSize-n2:len(serialized)-cryptography.PublicKeyHashHashSize], buf)
				copy(serialized[len(serialized)-cryptography.PublicKeyHashHashSize:], address.publicKeyHash)

				kernelHash := cryptography.SHA3Hash(serialized)

				kernelHash = cryptography.ComputeKernelHash(kernelHash, address.stakingAmount)

				if difficulty.CheckKernelHashBig(kernelHash, work.Target) {

					worker.workerSolutionCn <- &ForgingSolution{
						timestamp:     timestamp,
						address:       address.wallet,
						work:          work,
						stakingAmount: address.stakingAmount,
					}
					work = nil
					diff = 0
					break

				} else {
					// for debugging only
					//gui.Log(hex.EncodeToString(kernelHash), strconv.FormatUint(timestamp, 10 ))
				}

			}

			timestamp += 1
			timestampMs += 1000
		}
		atomic.AddUint32(&worker.hashes, uint32((diff+1)*len(wallets)))

	}

}

func createForgingWorkerThread(index int, workerSolutionCn chan *ForgingSolution) *ForgingWorkerThread {
	return &ForgingWorkerThread{
		index:            index,
		walletsCn:        make(chan []*ForgingWalletAddressRequired),
		workCn:           make(chan *forging_block_work.ForgingWork),
		workerSolutionCn: workerSolutionCn,
	}
}
