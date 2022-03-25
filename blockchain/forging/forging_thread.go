package forging

import (
	"bytes"
	"fmt"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"strconv"
	"sync/atomic"
	"time"
)

type ForgingThread struct {
	mempool                   *mempool.Mempool
	addressBalanceDecryptor   *address_balance_decryptor.AddressBalanceDecryptor
	threads                   int                                         //number of threads
	solutionCn                chan<- *blockchain_types.BlockchainSolution //broadcasting that a solution thread was received
	nextBlockCreatedCn        <-chan *forging_block_work.ForgingWork      //detect if a new work was published
	workers                   []*ForgingWorkerThread
	workersCreatedCn          chan []*ForgingWorkerThread
	workersDestroyedCn        chan struct{}
	lastPrevKernelHash        *generics.Value[[]byte]
	createForgingTransactions func(*block_complete.BlockComplete, []byte, uint64, []*transaction.Transaction) (*transaction.Transaction, error)
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
		thread.workers[i] = createForgingWorkerThread(i, forgingWorkerSolutionCn, thread.addressBalanceDecryptor)
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
		var newKernelHash []byte

		for {
			solution, ok := <-forgingWorkerSolutionCn
			if !ok {
				return
			}

			if solution.work.BlkComplete.Height > 1 && !bytes.Equal(solution.work.BlkComplete.PrevKernelHash, thread.lastPrevKernelHash.Load()) {
				return
			}

			if newKernelHash, err = thread.publishSolution(solution); err != nil {
				gui.GUI.Error(fmt.Errorf("Error publishing solution: %d error: %s ", solution.work.BlkHeight, err))
			} else {
				gui.GUI.Info(fmt.Errorf("Block was forged! %d ", solution.work.BlkHeight))
				thread.lastPrevKernelHash.Store(newKernelHash)
			}

		}
	})

	recovery.SafeGo(func() {
		for {
			newWork, ok := <-thread.nextBlockCreatedCn
			if !ok {
				return
			}

			thread.lastPrevKernelHash.Store(newWork.BlkComplete.PrevKernelHash)

			for i := 0; i < thread.threads; i++ {
				thread.workers[i].workCn <- newWork
			}

			gui.GUI.InfoUpdate("Hash Block", strconv.FormatUint(newWork.BlkHeight, 10))
		}
	})

}

func (thread *ForgingThread) publishSolution(solution *ForgingSolution) ([]byte, error) {

	work := solution.work

	newBlk := block_complete.CreateEmptyBlockComplete()
	if err := newBlk.Deserialize(helpers.NewBufferReader(work.BlkComplete.SerializeToBytes())); err != nil {
		return nil, err
	}

	newBlk.Block.StakingNonce = solution.stakingNonce
	newBlk.Block.Timestamp = solution.timestamp
	newBlk.Block.StakingAmount = solution.stakingAmount

	txs, _ := thread.mempool.GetNextTransactionsToInclude(newBlk.Block.PrevHash)

	txStakingReward, err := thread.createForgingTransactions(newBlk, solution.address.publicKey, solution.address.decryptedStakingBalance, txs)
	if err != nil {
		return nil, err
	}

	newBlk.Txs = append(txs, txStakingReward)

	newBlk.Block.MerkleHash = newBlk.MerkleHash()

	newBlk.Bloom = nil
	if err = newBlk.BloomAll(); err != nil {
		return nil, err
	}

	//send message to blockchain
	result := make(chan *blockchain_types.BlockchainSolutionAnswer)
	thread.solutionCn <- &blockchain_types.BlockchainSolution{
		newBlk,
		result,
	}

	res := <-result
	return res.ChainKernelHash, res.Err
}

func createForgingThread(threads int, createForgingTransactions func(*block_complete.BlockComplete, []byte, uint64, []*transaction.Transaction) (*transaction.Transaction, error), mempool *mempool.Mempool, addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor, solutionCn chan<- *blockchain_types.BlockchainSolution, nextBlockCreatedCn <-chan *forging_block_work.ForgingWork) *ForgingThread {
	return &ForgingThread{
		mempool,
		addressBalanceDecryptor,
		threads,
		solutionCn,
		nextBlockCreatedCn,
		[]*ForgingWorkerThread{},
		make(chan []*ForgingWorkerThread),
		make(chan struct{}),
		&generics.Value[[]byte]{},
		createForgingTransactions,
	}
}
