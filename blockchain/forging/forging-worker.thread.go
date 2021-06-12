package forging

import (
	"encoding/binary"
	difficulty "pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/config"
	"pandora-pay/config/stake"
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

type ForgingWorkerThread struct {
	hashes                uint32
	index                 int
	workCn                chan *forging_block_work.ForgingWork
	workerSolutionCn      chan *ForgingSolution
	addWalletAddressCn    chan *ForgingWalletAddress
	removeWalletAddressCn chan *ForgingWalletAddress
}

type ForgingWorkerThreadAddress struct {
	walletAdr     *ForgingWalletAddress
	stakingAmount uint64
}

func (threadAddr *ForgingWorkerThreadAddress) computeStakingAmount(height uint64) (err error) {

	threadAddr.stakingAmount = 0
	if threadAddr.walletAdr.account != nil && threadAddr.walletAdr.delegatedPrivateKey != nil {

		stakingAmount := uint64(0)
		if threadAddr.walletAdr.account != nil {
			if stakingAmount, err = threadAddr.walletAdr.account.ComputeDelegatedStakeAvailable(height); err != nil {
				return
			}
		}

		if stakingAmount >= stake.GetRequiredStake(height) {
			threadAddr.stakingAmount = stakingAmount
		}

	}
	return
}

/**
"Staking multiple wallets simultaneously"
*/
func (worker *ForgingWorkerThread) forge() {

	var work *forging_block_work.ForgingWork

	var timestampMs int64
	var timestamp, blkHeight uint64
	var serialized []byte
	var n int
	buf := make([]byte, binary.MaxVarintLen64)

	wallets := make(map[string]*ForgingWorkerThreadAddress)
	walletsStakable := make(map[string]*ForgingWorkerThreadAddress)

	for {

		timeLimitMs := time.Now().UnixNano()/1000000 + config.NETWORK_TIMESTAMP_DRIFT_MAX_INT*1000
		timeLimit := uint64(timeLimitMs / 1000)

		select {
		case newWork, ok := <-worker.workCn: //or the work was changed meanwhile
			if !ok {
				return
			}
			work = newWork

			serialized = helpers.CloneBytes(newWork.BlkSerialized)

			blkHeight = work.BlkHeight
			timestamp = work.BlkTimestmap + 1
			timestampMs = int64(timestamp) * 1000

			n = binary.PutUvarint(buf, timestamp)

			for _, walletAddr := range wallets {
				walletAddr.computeStakingAmount(blkHeight)
				if walletAddr.stakingAmount > 0 {
					walletsStakable[walletAddr.walletAdr.publicKeyHashStr] = walletAddr
				}
			}

		case newWalletAddress, ok := <-worker.addWalletAddressCn:
			if !ok {
				return
			}

			if wallets[string(newWalletAddress.publicKeyHash)] == nil {
				walletAddr := &ForgingWorkerThreadAddress{ //making sure i have a copy
					&ForgingWalletAddress{
						newWalletAddress.delegatedPrivateKey,
						newWalletAddress.delegatedPublicKeyHash,
						newWalletAddress.publicKeyHash,
						newWalletAddress.publicKeyHashStr,
						newWalletAddress.account,
						-1,
					},
					0,
				}
				wallets[newWalletAddress.publicKeyHashStr] = walletAddr
			}
			walletAddr := wallets[newWalletAddress.publicKeyHashStr]
			walletAddr.computeStakingAmount(blkHeight)
			if walletAddr.stakingAmount > 0 {
				walletsStakable[walletAddr.walletAdr.publicKeyHashStr] = walletAddr
			}
		case newWalletAddress, ok := <-worker.removeWalletAddressCn:
			if !ok {
				return
			}
			if wallets[newWalletAddress.publicKeyHashStr] != nil {
				delete(wallets, newWalletAddress.publicKeyHashStr)
				delete(walletsStakable, newWalletAddress.publicKeyHashStr)
			}
		default:
		}

		if work == nil || len(walletsStakable) == 0 {
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
				copy(serialized[len(serialized)-cryptography.PublicKeyHashHashSize:], address.walletAdr.publicKeyHash)

				kernelHash := cryptography.SHA3Hash(serialized)

				kernelHash = cryptography.ComputeKernelHash(kernelHash, address.stakingAmount)

				if difficulty.CheckKernelHashBig(kernelHash, work.Target) {

					select {
					default:
						worker.workerSolutionCn <- &ForgingSolution{
							timestamp:     timestamp,
							address:       address.walletAdr,
							work:          work,
							stakingAmount: address.stakingAmount,
						}
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
		index:                 index,
		workCn:                make(chan *forging_block_work.ForgingWork),
		workerSolutionCn:      workerSolutionCn,
		addWalletAddressCn:    make(chan *ForgingWalletAddress),
		removeWalletAddressCn: make(chan *ForgingWalletAddress),
	}
}
