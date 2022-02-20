package forging

import (
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/config/config_nodes"
	"pandora-pay/gui"
	"pandora-pay/helpers"
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

func (thread *ForgingThread) stopForging() {
	thread.workersDestroyedCn <- struct{}{}
	for i := 0; i < len(thread.workers); i++ {
		close(thread.workers[i].workCn)
	}
}

func (thread *ForgingThread) startForging() {

	thread.workers = make([]*ForgingWorkerThread, thread.threads)

	forgingWorkerSolutionCn := make(chan *ForgingSolution)
	for i := 0; i < len(thread.workers); i++ {
		thread.workers[i] = createForgingWorkerThread(i, forgingWorkerSolutionCn)
		recovery.SafeGo(thread.workers[i].forge)
	}
	thread.workersCreatedCn <- thread.workers

	recovery.SafeGo(func() {
		for {

			s := ""
			for i := 0; i < thread.threads; i++ {
				hashesPerSecond := atomic.SwapUint32(&thread.workers[i].hashes, 0)
				s += strconv.FormatUint(uint64(hashesPerSecond), 10) + " "
			}
			gui.GUI.InfoUpdate("Hashes/s", s)

			time.Sleep(time.Second)
		}
	})

	recovery.SafeGo(func() {
		var err error
		for {
			solution, ok := <-forgingWorkerSolutionCn
			if !ok {
				return
			}

			if err = thread.publishSolution(solution); err != nil {
				gui.GUI.Error("Error publishing solution", err)
			}
		}
	})

	recovery.SafeGo(func() {
		for {
			newWork, ok := <-thread.nextBlockCreatedCn
			if !ok {
				return
			}

			for i := 0; i < thread.threads; i++ {
				thread.workers[i].workCn <- newWork
			}
			gui.GUI.InfoUpdate("Hash Block", strconv.FormatUint(newWork.BlkHeight, 10))
		}
	})

}

func (thread *ForgingThread) publishSolution(solution *ForgingSolution) (err error) {

	work := solution.work

	newBlk := block_complete.CreateEmptyBlockComplete()
	if err = newBlk.Deserialize(helpers.NewBufferReader(work.BlkComplete.SerializeToBytes())); err != nil {
		return
	}

	newBlk.Block.StakingNonce = solution.stakingNonce
	newBlk.Block.Timestamp = solution.timestamp

	newBlk.Block.StakingAmount = solution.stakingAmount

	newBlk.Txs, _ = thread.mempool.GetNextTransactionsToInclude(newBlk.Block.PrevHash)
	newBlk.Block.MerkleHash = newBlk.MerkleHash()
	newBlk.Block.StakingFee = config_nodes.DELEGATOR_FEE

	newBlk.Bloom = nil
	if err = newBlk.BloomAll(); err != nil {
		return
	}

	//send message to blockchain
	thread.solutionCn <- newBlk

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
