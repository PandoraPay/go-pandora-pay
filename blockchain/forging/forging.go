package forging

import (
	"math/big"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
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

type Forging struct {
	work     unsafe.Pointer
	solution unsafe.Pointer

	started int32

	wg sync.WaitGroup

	mempool         *mempool.Mempool
	Wallet          *ForgingWallet
	SolutionChannel chan *block_complete.BlockComplete
}

func ForgingInit(mempool *mempool.Mempool) (forging *Forging, err error) {

	forging = &Forging{

		mempool:         mempool,
		SolutionChannel: make(chan *block_complete.BlockComplete),
		Wallet: &ForgingWallet{
			addressesMap: make(map[string]*ForgingWalletAddress),
		},
	}

	gui.Log("Forging Init")
	if err = forging.Wallet.loadBalances(); err != nil {
		return
	}

	go forging.startForging(config.CPU_THREADS)

	return
}

func (forging *Forging) startForging(threads int) {

	if !atomic.CompareAndSwapInt32(&forging.started, 0, 1) {
		return
	}

	var err error
	for atomic.LoadInt32(&forging.started) == 1 {

		workPointer := atomic.LoadPointer(&forging.work)
		if atomic.LoadPointer(&forging.work) == nil {
			// gui.Error("No block for staking..." )
			time.Sleep(10 * time.Millisecond)
			continue
		}
		work := (*ForgingWork)(workPointer)

		//distributing the wallets to each thread uniformly
		forging.Wallet.RLock()
		wallets := make([][]*ForgingWalletAddressRequired, threads)
		for i := 0; i < threads; i++ {
			wallets[i] = []*ForgingWalletAddressRequired{}
		}
		c := 0
		for i, walletAdr := range forging.Wallet.addresses {
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
					wallets[c%threads] = append(wallets[i%threads], &ForgingWalletAddressRequired{
						publicKeyHash: walletAdr.publicKeyHash,
						wallet:        walletAdr,
						stakingAmount: stakingAmount,
					})
					c++
				}
			}
		}
		forging.Wallet.RUnlock()

		for i := 0; i < threads; i++ {
			forging.wg.Add(1)
			go forge(forging, workPointer, work, wallets[i])
		}
		forging.wg.Wait()

		if atomic.LoadPointer(&forging.solution) != nil {
			err := forging.publishSolution()
			if err != nil {
				gui.Error("Error publishing solution", err)
				atomic.StorePointer(&forging.solution, nil)
			}
		}

	}

}

func (forging *Forging) StopForging() {
	atomic.StorePointer(&forging.work, nil)
	atomic.StorePointer(&forging.solution, nil)
	atomic.CompareAndSwapInt32(&forging.started, 1, 0)
}

//thread safe
func (forging *Forging) RestartForgingWorkers(blkComplete *block_complete.BlockComplete, target *big.Int) {

	work := ForgingWork{
		blkComplete: blkComplete,
		target:      target,
	}

	atomic.StorePointer(&forging.solution, nil)
	atomic.StorePointer(&forging.work, unsafe.Pointer(&work))

}

func (forging *Forging) StopForgingWorkers() {
	atomic.StorePointer(&forging.work, nil)
	atomic.StorePointer(&forging.solution, nil)
}

//thread safe
func (forging *Forging) foundSolution(address *ForgingWalletAddress, timestamp uint64, work *ForgingWork) {

	solution := ForgingSolution{
		timestamp: timestamp,
		address:   address,
		work:      work,
	}
	atomic.StorePointer(&forging.solution, unsafe.Pointer(&solution))
	atomic.StorePointer(&forging.work, nil)
}

// thread not safe
func (forging *Forging) publishSolution() (err error) {

	solutionPointer := atomic.LoadPointer(&forging.solution)
	solution := (*ForgingSolution)(solutionPointer)

	work := solution.work

	work.blkComplete.Block.Forger = solution.address.publicKeyHash
	work.blkComplete.Block.DelegatedPublicKeyHash = solution.address.delegatedPublicKeyHash
	work.blkComplete.Block.Timestamp = solution.timestamp

	if work.blkComplete.Block.Height > 0 {
		if work.blkComplete.Block.StakingAmount, err = solution.address.account.GetDelegatedStakeAvailable(work.blkComplete.Block.Height); err != nil {
			return
		}
	}

	work.blkComplete.Txs = forging.mempool.GetNextTransactionsToInclude(work.blkComplete.Block.Height, work.blkComplete.Block.PrevHash)
	work.blkComplete.Block.MerkleHash = work.blkComplete.MerkleHash()

	serializationForSigning := work.blkComplete.Block.SerializeForSigning()

	if work.blkComplete.Block.Signature, err = solution.address.delegatedPrivateKey.Sign(serializationForSigning); err != nil {
		return
	}

	//send message to blockchain
	forging.SolutionChannel <- work.blkComplete
	return
}

func (forging *Forging) Close() {
	forging.StopForging()
}
