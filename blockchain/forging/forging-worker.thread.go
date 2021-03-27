package forging

import (
	"encoding/binary"
	"math/big"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"sync/atomic"
	"time"
)

type ForgingWork struct {
	blkComplete *block_complete.BlockComplete
	target      *big.Int
}

type ForgingSolution struct {
	timestamp uint64
	address   *ForgingWalletAddress
	work      *ForgingWork
}
type ForgingWalletAddressRequired struct {
	publicKeyHash []byte //20 byte
	wallet        *ForgingWalletAddress
	stakingAmount uint64
}

type ForgingWorkerThread struct {
	hashes          uint32
	index           int
	workChannel     chan *ForgingWork                    // SAFE
	solutionChannel chan *ForgingSolution                // SAFE
	walletsChannel  chan []*ForgingWalletAddressRequired // SAFE
}

/**
"Staking multiple wallets simulateneoly"
*/
func (worker *ForgingWorkerThread) forge() {

	var work *ForgingWork
	var wallets []*ForgingWalletAddressRequired

	buf := make([]byte, binary.MaxVarintLen64)
	n := 0

	var height, timestamp uint64
	var serialized []byte

	for {

		timeNow := uint64(time.Now().Unix()) + config.NETWORK_TIMESTAMP_DRIFT_MAX

		select {
		case newWork, ok := <-worker.workChannel: //or the work was changed meanwhile
			if !ok {
				return
			}
			work = newWork
			height = work.blkComplete.Block.Height
			serialized = work.blkComplete.Block.SerializeForForging()
			n := binary.PutUvarint(buf, work.blkComplete.Block.Timestamp)

			serialized = serialized[:len(serialized)-n-20]
			timestamp = work.blkComplete.Block.Timestamp + 1
			atomic.StoreUint32(&worker.hashes, 0)
		case newWallets := <-worker.walletsChannel:
			wallets = newWallets
		default:
			if timestamp > timeNow {
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}

		if work == nil {
			time.Sleep(10 * time.Millisecond)
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
				serialized = append(serialized, buf[:n]...)
				serialized = append(serialized, address.publicKeyHash...)
				kernelHash := cryptography.SHA3Hash(serialized)

				if height > 0 {
					kernelHash = cryptography.ComputeKernelHash(kernelHash, address.stakingAmount)
				}

				if difficulty.CheckKernelHashBig(kernelHash, work.target) {

					worker.solutionChannel <- &ForgingSolution{
						timestamp: timestamp,
						address:   address.wallet,
						work:      work,
					}
					work = nil
					diff = 0
					break

				} else {
					//for debugging only
					//gui.Log(hex.EncodeToString(kernelHash))
				}

				serialized = serialized[:len(serialized)-n-20]
			}

			timestamp += 1
		}
		atomic.AddUint32(&worker.hashes, uint32((diff+1)*len(wallets)))

	}

}

func createForgingWorkerThread(index int, solutionChannel chan *ForgingSolution) *ForgingWorkerThread {
	return &ForgingWorkerThread{
		index:           index,
		walletsChannel:  make(chan []*ForgingWalletAddressRequired),
		workChannel:     make(chan *ForgingWork),
		solutionChannel: solutionChannel,
	}
}