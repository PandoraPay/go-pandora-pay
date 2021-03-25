package forging

import (
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config/globals"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"strconv"
	"sync/atomic"
	"time"
)

type ForgingThread struct {
	mempool         *mempool.Mempool
	threads         int                                  //number of threads
	wallet          *ForgingWallet                       //shared wallet, not thread safe
	solutionChannel chan<- *block_complete.BlockComplete //broadcasting that a solution thread was received
	workChannel     <-chan *ForgingWork                  //detect if a new work was published
}

func (thread *ForgingThread) getWallets(wallet *ForgingWallet, work *ForgingWork) [][]*ForgingWalletAddressRequired {

	var err error
	wallets := make([][]*ForgingWalletAddressRequired, thread.threads)

	//distributing the wallets to each thread uniformly
	wallet.RLock()
	for i := 0; i < thread.threads; i++ {
		wallets[i] = []*ForgingWalletAddressRequired{}
	}
	c := 0
	for i, walletAdr := range wallet.addresses {
		if walletAdr.account != nil || work.blkComplete.Block.Height == 0 {

			if work.blkComplete.Block.Height == 0 && i > 0 && globals.Arguments["--new-devnet"] == true {
				break
			}

			var stakingAmount uint64
			if walletAdr.account != nil {
				stakingAmount, err = walletAdr.account.GetDelegatedStakeAvailable(work.blkComplete.Block.Height)
				if err != nil {
					continue
				}
			}
			if stakingAmount >= stake.GetRequiredStake(work.blkComplete.Block.Height) {
				wallets[c%thread.threads] = append(wallets[c%thread.threads], &ForgingWalletAddressRequired{
					publicKeyHash: walletAdr.delegatedPublicKeyHash,
					wallet:        walletAdr,
					stakingAmount: stakingAmount,
				})
				c++
			}
		}
	}
	wallet.RUnlock()
	return wallets
}

func (thread *ForgingThread) startForging() {

	workers := make([]*ForgingWorkerThread, thread.threads)
	forgingWorkerSolutionChannel := make(chan *ForgingSolution, 0)
	for i := 0; i < len(workers); i++ {
		workers[i] = createForgingWorkerThread(i, forgingWorkerSolutionChannel)
		go workers[i].forge()
	}

	//wallets must be read only after its assignment
	ticker := time.NewTicker(1 * time.Second)

	defer func() {
		for i := 0; i < len(workers); i++ {
			close(workers[i].workChannel)
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
	for {

		work, ok := <-thread.workChannel
		if !ok {
			return
		}

		wallets := thread.getWallets(thread.wallet, work)

		for i := 0; i < thread.threads; i++ {
			workers[i].walletsChannel <- wallets[i]
		}
		for i := 0; i < thread.threads; i++ {
			workers[i].workChannel <- work
		}

		select {
		case solution := <-forgingWorkerSolutionChannel:
			if err = thread.publishSolution(solution); err != nil {
				gui.Error("Error publishing solution", err)
			}
			break
		case work, ok = <-thread.workChannel:
			if !ok {
				return
			}
			break //it was changed
		}

	}

}

func (thread *ForgingThread) publishSolution(solution *ForgingSolution) (err error) {

	work := solution.work

	work.blkComplete.Block.Forger = solution.address.publicKeyHash
	work.blkComplete.Block.Timestamp = solution.timestamp

	if work.blkComplete.Block.Height > 0 {
		if work.blkComplete.Block.StakingAmount, err = solution.address.account.GetDelegatedStakeAvailable(work.blkComplete.Block.Height); err != nil {
			return
		}
	}

	work.blkComplete.Txs = thread.mempool.GetNextTransactionsToInclude(work.blkComplete.Block.Height, work.blkComplete.Block.PrevHash)
	work.blkComplete.Block.MerkleHash = work.blkComplete.MerkleHash()

	hashForSignature := work.blkComplete.Block.SerializeForSigning()

	if work.blkComplete.Block.Signature, err = solution.address.delegatedPrivateKey.Sign(hashForSignature); err != nil {
		return
	}

	//send message to blockchain
	thread.solutionChannel <- work.blkComplete
	return
}

func createForgingThread(threads int, mempool *mempool.Mempool, solutionChannel chan<- *block_complete.BlockComplete, workChannel <-chan *ForgingWork, wallet *ForgingWallet) *ForgingThread {
	return &ForgingThread{
		threads:         threads,
		mempool:         mempool,
		solutionChannel: solutionChannel,
		workChannel:     workChannel,
		wallet:          wallet,
	}
}
