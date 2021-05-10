package forging

import (
	"encoding/binary"
	"math/big"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"sync/atomic"
	"time"
)

type ForgingWork struct {
	blkComplete *block_complete.BlockComplete
	target      *big.Int
}

type ForgingSolution struct {
	timestamp     uint64
	address       *ForgingWalletAddress
	work          *ForgingWork
	stakingAmount uint64
}
type ForgingWalletAddressRequired struct {
	publicKeyHash []byte //20 byte
	wallet        *ForgingWalletAddress
	stakingAmount uint64
}

type ForgingWorkerThread struct {
	hashes           uint32
	index            int
	workCn           chan *ForgingWork                    // SAFE
	workerSolutionCn chan *ForgingSolution                // SAFE
	walletsCn        chan []*ForgingWalletAddressRequired // SAFE
}

/**
"Staking multiple wallets simulateneoly"
*/
func (worker *ForgingWorkerThread) forge() {

	var work *ForgingWork
	var wallets []*ForgingWalletAddressRequired

	buf := make([]byte, binary.MaxVarintLen64)
	n := 0

	var timestamp uint64
	var serialized []byte

	for {

		timeNow := uint64(time.Now().Unix()) + config.NETWORK_TIMESTAMP_DRIFT_MAX

		select {
		case newWork, ok := <-worker.workCn: //or the work was changed meanwhile
			if !ok {
				return
			}
			work = newWork

			writer := helpers.NewBufferWriter()
			work.blkComplete.Block.SerializeForForging(writer)
			serialized = writer.Bytes()

			n := binary.PutUvarint(buf, work.blkComplete.Block.Timestamp)

			serialized = serialized[:len(serialized)-n-20]
			timestamp = work.blkComplete.Block.Timestamp + 1
			atomic.StoreUint32(&worker.hashes, 0)
		case newWallets := <-worker.walletsCn:
			wallets = newWallets
		default:
		}

		if work == nil || timestamp > timeNow {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		//forge with my wallets
		diff := int(timeNow - timestamp)
		if diff > 20 {
			diff = 20
		}
		for i := 0; i <= diff; i++ {
			for _, address := range wallets {

				n = binary.PutUvarint(buf, timestamp)

				final := append(serialized, buf[:n]...)
				final = append(final, address.publicKeyHash...)
				kernelHash := cryptography.SHA3Hash(final)

				kernelHash = cryptography.ComputeKernelHash(kernelHash, address.stakingAmount)

				if difficulty.CheckKernelHashBig(kernelHash, work.target) {

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
					//for debugging only
					//gui.Log(hex.EncodeToString(kernelHash))
				}

			}

			timestamp += 1
		}
		atomic.AddUint32(&worker.hashes, uint32((diff+1)*len(wallets)))

	}

}

func createForgingWorkerThread(index int, workerSolutionCn chan *ForgingSolution) *ForgingWorkerThread {
	return &ForgingWorkerThread{
		index:            index,
		walletsCn:        make(chan []*ForgingWalletAddressRequired),
		workCn:           make(chan *ForgingWork),
		workerSolutionCn: workerSolutionCn,
	}
}
