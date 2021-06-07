package forging

import (
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"strconv"
	"sync/atomic"
	"time"
)

type ForgingThread struct {
	mempool            *mempool.Mempool
	threads            int                                    //number of threads
	wallet             *ForgingWallet                         //shared wallet, not thread safe
	solutionCn         chan<- *block_complete.BlockComplete   //broadcasting that a solution thread was received
	nextBlockCreatedCn <-chan *forging_block_work.ForgingWork //detect if a new work was published
}

func (thread *ForgingThread) getWallets(wallet *ForgingWallet, work *forging_block_work.ForgingWork) (wallets [][]*ForgingWalletAddressRequired, walletsCount int) {

	wallets = make([][]*ForgingWalletAddressRequired, thread.threads)

	//distributing the wallets to each thread uniformly
	wallet.RLock()
	defer wallet.RUnlock()
	for i := 0; i < thread.threads; i++ {
		wallets[i] = []*ForgingWalletAddressRequired{}
	}

	walletsCount = 0
	for _, walletAdr := range wallet.addresses {
		if walletAdr.account != nil && walletAdr.delegatedPrivateKey != nil {

			stakingAmount := uint64(0)
			if walletAdr.account != nil {
				var err error
				if stakingAmount, err = walletAdr.account.ComputeDelegatedStakeAvailable(work.BlkComplete.Block.Height); err != nil {
					continue
				}
			}

			if stakingAmount >= stake.GetRequiredStake(work.BlkComplete.Block.Height) {
				wallets[walletsCount%thread.threads] = append(wallets[walletsCount%thread.threads], &ForgingWalletAddressRequired{
					publicKeyHash: walletAdr.publicKeyHash,
					wallet:        walletAdr,
					stakingAmount: stakingAmount,
				})
				walletsCount++
			}

		}
	}
	return
}

func (thread *ForgingThread) startForging() {

	workers := make([]*ForgingWorkerThread, thread.threads)
	forgingWorkerSolutionCn := make(chan *ForgingSolution, 0)
	for i := 0; i < len(workers); i++ {
		workers[i] = createForgingWorkerThread(i, forgingWorkerSolutionCn)
		recovery.SafeGo(workers[i].forge)
	}

	//wallets must be read only after its assignment
	ticker := time.NewTicker(1 * time.Second)

	defer func() {
		for i := 0; i < len(workers); i++ {
			close(workers[i].workCn)
		}
		ticker.Stop()
	}()

	recovery.SafeGo(func() {
		for {

			if _, ok := <-ticker.C; !ok {
				return
			}

			s := ""
			for i := 0; i < thread.threads; i++ {
				hashesPerSecond := atomic.SwapUint32(&workers[i].hashes, 0)
				s += strconv.FormatUint(uint64(hashesPerSecond), 10) + " "
			}
			gui.GUI.InfoUpdate("Hashes/s", s)
		}
	})

	var err error
	var ok bool
	var work *forging_block_work.ForgingWork
	readNextWork := true
	for {

		if readNextWork {
			work, ok = <-thread.nextBlockCreatedCn
			if !ok {
				return
			}
		}
		readNextWork = true

		wallets, _ := thread.getWallets(thread.wallet, work)

		for i := 0; i < thread.threads; i++ {
			workers[i].walletsCn <- wallets[i]
		}
		for i := 0; i < thread.threads; i++ {
			workers[i].workCn <- work
		}

		select {
		case solution, ok := <-forgingWorkerSolutionCn:
			if !ok {
				return
			}
			if err = thread.publishSolution(solution); err != nil {
				gui.GUI.Error("Error publishing solution", err)
			}
		case work, ok = <-thread.nextBlockCreatedCn:
			if !ok {
				return
			}
			readNextWork = false
		}

	}

}

func (thread *ForgingThread) publishSolution(solution *ForgingSolution) (err error) {

	work := solution.work

	work.BlkComplete.Block.Forger = solution.address.publicKeyHash
	work.BlkComplete.Block.Timestamp = solution.timestamp

	work.BlkComplete.Block.StakingAmount = solution.stakingAmount

	work.BlkComplete.Txs = thread.mempool.GetNextTransactionsToInclude(work.BlkComplete.Block.Height, work.BlkComplete.Block.PrevHash)
	work.BlkComplete.Block.MerkleHash = work.BlkComplete.MerkleHash()

	hashForSignature := work.BlkComplete.Block.SerializeForSigning()

	if work.BlkComplete.Block.Signature, err = solution.address.delegatedPrivateKey.Sign(hashForSignature); err != nil {
		return
	}

	//send message to blockchain
	thread.solutionCn <- work.BlkComplete
	return
}

func createForgingThread(threads int, mempool *mempool.Mempool, solutionCn chan<- *block_complete.BlockComplete, nextBlockCreatedCn <-chan *forging_block_work.ForgingWork, wallet *ForgingWallet) *ForgingThread {
	return &ForgingThread{
		threads:            threads,
		mempool:            mempool,
		solutionCn:         solutionCn,
		nextBlockCreatedCn: nextBlockCreatedCn,
		wallet:             wallet,
	}
}
