package forging

import (
	"encoding/binary"
	difficulty "pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
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
	continueCn            chan struct{}
	workerSolutionCn      chan *ForgingSolution
	addWalletAddressCn    chan *ForgingWalletAddress
	removeWalletAddressCn chan string //publicKey
}

type ForgingWorkerThreadAddress struct {
	walletAdr     *ForgingWalletAddress
	stakingAmount uint64
}

func (threadAddr *ForgingWorkerThreadAddress) computeStakingAmount(height uint64) uint64 {

	threadAddr.stakingAmount = 0
	if threadAddr.walletAdr.account != nil && threadAddr.walletAdr.delegatedPrivateKey != nil {

		stakingAmount := uint64(0)
		if threadAddr.walletAdr.account != nil {
			stakingAmount, _ = threadAddr.walletAdr.account.ComputeDelegatedStakeAvailable(height)
		}

		if stakingAmount >= config_stake.GetRequiredStake(height) {
			threadAddr.stakingAmount = stakingAmount
			return stakingAmount
		}

	}
	return 0
}

/**
"Staking multiple wallets simultaneously"
*/
func (worker *ForgingWorkerThread) forge() {

	var work *forging_block_work.ForgingWork
	suspended := true

	var timestampMs int64
	var timestamp, blkHeight uint64
	var serialized []byte
	var n int
	buf := make([]byte, binary.MaxVarintLen64)

	wallets := make(map[string]*ForgingWorkerThreadAddress)
	walletsStakable := make(map[string]*ForgingWorkerThreadAddress)

	waitCn := make(chan struct{})
	waitCnClosed := false

	validateWork := func() {
		if suspended || work == nil || len(walletsStakable) == 0 {
			if waitCnClosed {
				waitCn = make(chan struct{})
				waitCnClosed = false
			}
		} else {
			if !waitCnClosed {
				close(waitCn)
				waitCnClosed = true
			}
		}
	}

	for {

		select {
		case <-worker.continueCn:
			suspended = false
			validateWork()
		case newWork := <-worker.workCn: //or the work was changed meanwhile

			if newWork == nil {
				continue
			}

			work = newWork

			serialized = helpers.CloneBytes(work.BlkSerialized)

			blkHeight = work.BlkHeight
			timestamp = work.BlkTimestmap + 1
			timestampMs = int64(timestamp) * 1000

			n = binary.PutUvarint(buf, timestamp)

			walletsStakable = make(map[string]*ForgingWorkerThreadAddress)
			for _, walletAddr := range wallets {
				if walletAddr.computeStakingAmount(blkHeight) > 0 {
					walletsStakable[walletAddr.walletAdr.publicKeyStr] = walletAddr
				}
			}

			validateWork()
		case newWalletAddr := <-worker.addWalletAddressCn:
			walletAddr := wallets[newWalletAddr.publicKeyStr]
			if walletAddr == nil {
				walletAddr = &ForgingWorkerThreadAddress{ //making sure i have a copy
					newWalletAddr.clone(),
					0,
				}
				wallets[newWalletAddr.publicKeyStr] = walletAddr
			} else {
				walletAddr.walletAdr = newWalletAddr
			}
			if walletAddr.computeStakingAmount(blkHeight) > 0 {
				walletsStakable[walletAddr.walletAdr.publicKeyStr] = walletAddr
			} else {
				delete(walletsStakable, walletAddr.walletAdr.publicKeyStr)
			}
			validateWork()
		case publicKeyStr := <-worker.removeWalletAddressCn:
			if wallets[publicKeyStr] != nil {
				delete(wallets, publicKeyStr)
				delete(walletsStakable, publicKeyStr)
			}
			validateWork()
		case <-waitCn:
		}

		if !waitCnClosed {
			continue
		}

		timeLimitMs := time.Now().UnixNano()/1000000 + config.NETWORK_TIMESTAMP_DRIFT_MAX_INT*1000

		if timestampMs > timeLimitMs {
			time.Sleep(time.Millisecond * time.Duration(timestampMs-timeLimitMs))
			continue
		}

		timeLimit := uint64(timeLimitMs / 1000)
		//forge with my wallets
		diff := int(timeLimit - timestamp)
		if diff > 20 {
			diff = 20
		}
		if diff < 0 {
			diff = 0
		}

		for i := 0; i <= diff; i++ {
			for _, address := range walletsStakable {

				n2 := binary.PutUvarint(buf, timestamp)

				if n2 != n {
					newSerialized := make([]byte, len(serialized)-n+n2)
					copy(newSerialized, serialized[:-n-cryptography.PublicKeySize])
					serialized = newSerialized
					n = n2
				}

				//optimized POS
				copy(serialized[len(serialized)-cryptography.PublicKeySize-n2:len(serialized)-cryptography.PublicKeySize], buf)
				copy(serialized[len(serialized)-cryptography.PublicKeySize:], address.walletAdr.publicKey)

				kernelHash := cryptography.SHA3(serialized)

				kernelHash = cryptography.ComputeKernelHash(kernelHash, address.stakingAmount)

				if difficulty.CheckKernelHashBig(kernelHash, work.Target) {

					worker.workerSolutionCn <- &ForgingSolution{
						timestamp,
						address.walletAdr,
						work,
						address.stakingAmount,
					}

					suspended = true
					validateWork()

					diff = -1
					break

				} else {
					// for debugging only
					// gui.GUI.Log(hex.EncodeToString(kernelHash), strconv.FormatUint(timestamp, 10 ))
				}

			}

			timestamp += 1
			timestampMs += 1000
		}
		atomic.AddUint32(&worker.hashes, uint32((diff+1)*len(walletsStakable)))

	}

}

func createForgingWorkerThread(index int, workerSolutionCn chan *ForgingSolution) *ForgingWorkerThread {
	return &ForgingWorkerThread{
		index:                 index,
		continueCn:            make(chan struct{}),
		workCn:                make(chan *forging_block_work.ForgingWork),
		workerSolutionCn:      workerSolutionCn,
		addWalletAddressCn:    make(chan *ForgingWalletAddress),
		removeWalletAddressCn: make(chan string),
	}
}
