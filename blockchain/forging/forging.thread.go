package forging

import (
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config/globals"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"sync"
)

func startForging(
	mempool *mempool.Mempool,
	solutionChannel chan *block_complete.BlockComplete,
	workChannel <-chan *ForgingWork, //detect if a new work was published
	wallet *ForgingWallet, //shared wallet, not thread safe
	threads int, //number of threads
) {

	wg := sync.WaitGroup{}

	//wallets must be read only after its assignment
	wallets := make([][]*ForgingWalletAddressRequired, threads)

	var err error
	for {

		work, ok := <-workChannel
		if !ok {
			return
		}

		//distributing the wallets to each thread uniformly
		wallet.RLock()
		for i := 0; i < threads; i++ {
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
					wallets[c%threads] = append(wallets[c%threads], &ForgingWalletAddressRequired{
						publicKeyHash: walletAdr.delegatedPublicKeyHash,
						wallet:        walletAdr,
						stakingAmount: stakingAmount,
					})
					c++
				}
			}
		}
		wallet.RUnlock()

		stakingSolutionChannel := make(chan *ForgingSolution, 0)
		for i := 0; i < threads; i++ {
			wg.Add(1)
			go forge(&wg, work, workChannel, stakingSolutionChannel, wallets[i])
		}

		solution := <-stakingSolutionChannel

		select {
		case work, ok = <-workChannel:
			if !ok {
				return
			}
			break //it was changed
		default:
			if err = publishSolution(mempool, solutionChannel, solution); err != nil {
				gui.Error("Error publishing solution", err)
			}
		}

	}

}

func publishSolution(mempool *mempool.Mempool, solutionChannel chan *block_complete.BlockComplete, solution *ForgingSolution) (err error) {

	work := solution.work

	work.blkComplete.Block.Forger = solution.address.publicKeyHash
	work.blkComplete.Block.Timestamp = solution.timestamp

	if work.blkComplete.Block.Height > 0 {
		if work.blkComplete.Block.StakingAmount, err = solution.address.account.GetDelegatedStakeAvailable(work.blkComplete.Block.Height); err != nil {
			return
		}
	}

	work.blkComplete.Txs = mempool.GetNextTransactionsToInclude(work.blkComplete.Block.Height, work.blkComplete.Block.PrevHash)
	work.blkComplete.Block.MerkleHash = work.blkComplete.MerkleHash()

	hashForSignature := work.blkComplete.Block.SerializeForSigning()

	if work.blkComplete.Block.Signature, err = solution.address.delegatedPrivateKey.Sign(hashForSignature); err != nil {
		return
	}

	//send message to blockchain
	solutionChannel <- work.blkComplete
	return
}
