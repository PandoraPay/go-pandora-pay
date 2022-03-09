package forging

import (
	"github.com/tevino/abool"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
)

type Forging struct {
	mempool                 *mempool.Mempool
	addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor
	Wallet                  *ForgingWallet
	started                 *abool.AtomicBool
	forgingThread           *ForgingThread
	nextBlockCreatedCn      <-chan *forging_block_work.ForgingWork
	forgingSolutionCn       chan<- *block_complete.BlockComplete
}

func CreateForging(mempool *mempool.Mempool, addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor) (*Forging, error) {

	forging := &Forging{
		mempool,
		addressBalanceDecryptor,
		&ForgingWallet{
			addressBalanceDecryptor,
			map[string]*ForgingWalletAddress{},
			[]int{},
			[]*ForgingWorkerThread{},
			nil,
			make(chan *ForgingWalletAddressUpdate),
			nil,
			nil,
			&generics.Map[string, *ForgingWalletAddress]{},
			nil,
		},
		abool.New(),
		nil, nil, nil,
	}
	forging.Wallet.forging = forging

	return forging, nil
}

func (forging *Forging) InitializeForging(createForgingTransactions func(*block_complete.BlockComplete, []byte, uint64, []*transaction.Transaction) (*transaction.Transaction, error), nextBlockCreatedCn <-chan *forging_block_work.ForgingWork, updateNewChainUpdate *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates], forgingSolutionCn chan<- *block_complete.BlockComplete) {

	forging.nextBlockCreatedCn = nextBlockCreatedCn
	forging.Wallet.updateNewChainUpdate = updateNewChainUpdate
	forging.forgingSolutionCn = forgingSolutionCn

	forging.forgingThread = createForgingThread(config.CPU_THREADS, createForgingTransactions, forging.mempool, forging.addressBalanceDecryptor, forging.forgingSolutionCn, forging.nextBlockCreatedCn)
	forging.Wallet.workersCreatedCn = forging.forgingThread.workersCreatedCn
	forging.Wallet.workersDestroyedCn = forging.forgingThread.workersDestroyedCn

	recovery.SafeGo(forging.Wallet.runProcessUpdates)
	recovery.SafeGo(forging.Wallet.runDecryptBalanceAndNotifyWorkers)

}

func (forging *Forging) StartForging() bool {

	if config.CONSENSUS != config.CONSENSUS_TYPE_FULL {
		gui.GUI.Warning(`Staking was not started as "--consensus=full" is missing`)
		return false
	}

	if !forging.started.SetToIf(false, true) {
		return false
	}

	forging.forgingThread.startForging()

	return true
}

func (forging *Forging) StopForging() bool {
	if forging.started.SetToIf(true, false) {
		return true
	}
	return false
}

func (forging *Forging) Close() {
	forging.StopForging()
}
