package forging

import (
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"strconv"
	"sync/atomic"
	"time"
)

type ForgingThread struct {
	mempool    *mempool.Mempool
	threads    int                                  //number of threads
	wallet     *ForgingWallet                       //shared wallet, not thread safe
	solutionCn chan<- *block_complete.BlockComplete //broadcasting that a solution thread was received
	workCn     <-chan *ForgingWork                  //detect if a new work was published
}

func (thread *ForgingThread) getWallets(wallet *ForgingWallet, work *ForgingWork) (wallets [][]*ForgingWalletAddressRequired, walletsCount int) {

	wallets = make([][]*ForgingWalletAddressRequired, thread.threads)

	//distributing the wallets to each thread uniformly
	wallet.RLock()
	for i := 0; i < thread.threads; i++ {
		wallets[i] = []*ForgingWalletAddressRequired{}
	}

	walletsCount = 0
	for _, walletAdr := range wallet.addresses {
		if walletAdr.account != nil && walletAdr.delegatedPrivateKey != nil {

			stakingAmount := uint64(0)
			if walletAdr.account != nil {
				var err error
				if stakingAmount, err = walletAdr.account.ComputeDelegatedStakeAvailable(work.blkComplete.Block.Height); err != nil {
					continue
				}
			}

			if stakingAmount >= stake.GetRequiredStake(work.blkComplete.Block.Height) {
				wallets[walletsCount%thread.threads] = append(wallets[walletsCount%thread.threads], &ForgingWalletAddressRequired{
					publicKeyHash: walletAdr.publicKeyHash,
					wallet:        walletAdr,
					stakingAmount: stakingAmount,
				})
				walletsCount++
			}

		}
	}
	wallet.RUnlock()
	return
}

func (thread *ForgingThread) startForging() {

	workers := make([]*ForgingWorkerThread, thread.threads)
	forgingWorkerSolutionCn := make(chan *ForgingSolution, 0)
	for i := 0; i < len(workers); i++ {
		workers[i] = createForgingWorkerThread(i, forgingWorkerSolutionCn)
		go workers[i].forge()
	}

	//wallets must be read only after its assignment
	ticker := time.NewTicker(1 * time.Second)

	defer func() {
		for i := 0; i < len(workers); i++ {
			close(workers[i].workCn)
		}
		ticker.Stop()
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				s := ""
				for i := 0; i < thread.threads; i++ {
					hashesPerSecond := atomic.SwapUint32(&workers[i].hashes, 0)
					s += strconv.FormatUint(uint64(hashesPerSecond), 10) + " "
				}
				gui.InfoUpdate("Hashes/s", s)
			}
		}
	}()

	var err error
	var ok bool
	var work *ForgingWork
	readNextWork := true
	for {

		if readNextWork {
			work, ok = <-thread.workCn
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
		case solution := <-forgingWorkerSolutionCn:
			if err = thread.publishSolution(solution); err != nil {
				gui.Error("Error publishing solution", err)
			}
			break
		case work, ok = <-thread.workCn:
			if !ok {
				return
			}
			readNextWork = false
			break //it was changed
		}

	}

}

func (thread *ForgingThread) publishSolution(solution *ForgingSolution) (err error) {

	work := solution.work

	work.blkComplete.Block.Forger = solution.address.publicKeyHash
	work.blkComplete.Block.Timestamp = solution.timestamp

	work.blkComplete.Block.StakingAmount = solution.stakingAmount

	work.blkComplete.Txs = thread.mempool.GetNextTransactionsToInclude(work.blkComplete.Block.Height, work.blkComplete.Block.PrevHash)
	work.blkComplete.Block.MerkleHash = work.blkComplete.MerkleHash()

	hashForSignature := work.blkComplete.Block.SerializeForSigning()

	if work.blkComplete.Block.Signature, err = solution.address.delegatedPrivateKey.Sign(hashForSignature); err != nil {
		return
	}

	//send message to blockchain
	thread.solutionCn <- work.blkComplete
	return
}

func createForgingThread(threads int, mempool *mempool.Mempool, solutionCn chan<- *block_complete.BlockComplete, workCn <-chan *ForgingWork, wallet *ForgingWallet) *ForgingThread {
	return &ForgingThread{
		threads:    threads,
		mempool:    mempool,
		solutionCn: solutionCn,
		workCn:     workCn,
		wallet:     wallet,
	}
}
