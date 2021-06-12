package forging

import (
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/forging/forging-block-work"
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
	solutionCn         chan<- *block_complete.BlockComplete   //broadcasting that a solution thread was received
	nextBlockCreatedCn <-chan *forging_block_work.ForgingWork //detect if a new work was published
	workers            []*ForgingWorkerThread
	workersCreatedCn   chan []*ForgingWorkerThread
	workersDestroyedCn chan struct{}
}

func (thread *ForgingThread) startForging() {

	thread.workers = make([]*ForgingWorkerThread, thread.threads)

	forgingWorkerSolutionCn := make(chan *ForgingSolution)
	for i := 0; i < len(thread.workers); i++ {
		thread.workers[i] = createForgingWorkerThread(i, forgingWorkerSolutionCn)
		recovery.SafeGo(thread.workers[i].forge)
	}
	thread.workersCreatedCn <- thread.workers

	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		thread.workersDestroyedCn <- struct{}{}
		for i := 0; i < len(thread.workers); i++ {
			close(thread.workers[i].workCn)
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
				hashesPerSecond := atomic.SwapUint32(&thread.workers[i].hashes, 0)
				s += strconv.FormatUint(uint64(hashesPerSecond), 10) + " "
			}
			gui.GUI.InfoUpdate("Hashes/s", s)
		}
	})

	var err error
	var ok bool
	var newWork *forging_block_work.ForgingWork
	for {

		select {
		case solution, ok := <-forgingWorkerSolutionCn:
			if !ok {
				return
			}
			if err = thread.publishSolution(solution); err != nil {
				gui.GUI.Error("Error publishing solution", err)
			}
		case newWork, ok = <-thread.nextBlockCreatedCn:
			if !ok {
				return
			}
			for i := 0; i < thread.threads; i++ {
				thread.workers[i].workCn <- newWork
			}
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

func createForgingThread(threads int, mempool *mempool.Mempool, solutionCn chan<- *block_complete.BlockComplete, nextBlockCreatedCn <-chan *forging_block_work.ForgingWork) *ForgingThread {
	return &ForgingThread{
		mempool,
		threads,
		solutionCn,
		nextBlockCreatedCn,
		[]*ForgingWorkerThread{},
		make(chan []*ForgingWorkerThread),
		make(chan struct{}),
	}
}
