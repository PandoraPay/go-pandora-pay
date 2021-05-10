package mempool

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/store"
	"sort"
	"sync/atomic"
	"time"
)

type mempoolWork struct {
	chainHash   []byte //32 byte
	chainHeight uint64
	result      *mempoolResult
}

type mempoolWorker struct {
	work        *mempoolWork
	workChanged bool
	boltTx      *bbolt.Tx
	accs        *accounts.Accounts
	toks        *tokens.Tokens
}

func (worker *mempoolWorker) closeDB() {
	if worker.boltTx != nil {
		worker.boltTx.Rollback()
		worker.boltTx = nil
	}
}

func sortTxs(txList []*mempoolTx) {
	sort.Slice(txList, func(i, j int) bool {

		if txList[i].FeePerByte == txList[j].FeePerByte && txList[i].Tx.TxType == transaction_type.TxSimple && txList[j].Tx.TxType == transaction_type.TxSimple {
			return txList[i].Tx.TxBase.(*transaction_simple.TransactionSimple).Nonce < txList[j].Tx.TxBase.(*transaction_simple.TransactionSimple).Nonce
		}

		return txList[i].FeePerByte < txList[j].FeePerByte
	})
}

//process the worker for transactions to prepare the transactions to the forger
func (worker *mempoolWorker) processing(
	newWork <-chan *mempoolWork, //SAFE
	mempoolTxs *mempoolTxs, //NOT SAFE, need RLOCK!
) {

	var txList []*mempoolTx
	var txMap map[string]bool

	listIndex := -1
	for {

		//let's check hf the work has been changed
		select {
		case work, ok := <-newWork:
			if !ok {
				return
			}

			if work != nil {
				worker.closeDB()
				worker.work = work
			}
			worker.workChanged = true
			txMap = make(map[string]bool)
			listIndex = -1

		default:

			if worker.work == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if listIndex == -1 {

				txListAll := mempoolTxs.txsList.Load().([]*mempoolTx)

				if worker.workChanged { //it is faster to copy first
					txList = txListAll
				} else {
					for _, tx := range txListAll {
						if !txMap[tx.Tx.Bloom.HashStr] {
							txList = append(txList, tx)
						}
					}
				}

				worker.workChanged = false

				if len(txList) > 0 {

					sortTxs(txList)

					var err error
					worker.boltTx, err = store.StoreBlockchain.DB.Begin(false)
					if err != nil {
						worker.closeDB()
						gui.Error("Error opening database for mempool")
						time.Sleep(1000 * time.Millisecond)
						continue
					}
					worker.accs = accounts.NewAccounts(worker.boltTx)
					worker.toks = tokens.NewTokens(worker.boltTx)
				}
				listIndex = 0

			} else {

				if listIndex == len(txList) {
					worker.closeDB()
					listIndex = -1
					time.Sleep(1000 * time.Millisecond)
					continue
				} else {

					tx := txList[listIndex]

					if txMap[tx.Tx.Bloom.HashStr] {
						listIndex += 1
						continue
					}

					txMap[tx.Tx.Bloom.HashStr] = true
					if err := txList[listIndex].Tx.IncludeTransaction(worker.work.chainHeight, worker.accs, worker.toks); err != nil {
						worker.accs.Rollback()
						worker.toks.Rollback()

						worker.work.result.txsErrorsMutex.Lock()
						txsErrors := worker.work.result.txsErrors.Load().([]*mempoolTx)
						worker.work.result.txsErrors.Store(append(txsErrors, tx))
						worker.work.result.txsErrorsMutex.Unlock()

					} else {
						totalSize := atomic.LoadUint64(&worker.work.result.totalSize)
						if totalSize+txList[listIndex].Tx.Bloom.Size < config.BLOCK_MAX_SIZE {
							worker.work.result.txsMutex.Lock()

							totalSize = atomic.LoadUint64(&worker.work.result.totalSize) + txList[listIndex].Tx.Bloom.Size
							if totalSize < config.BLOCK_MAX_SIZE {
								atomic.StoreUint64(&worker.work.result.totalSize, totalSize)
								txs := worker.work.result.txs.Load().([]*mempoolTx)
								worker.work.result.txs.Store(append(txs, txList[listIndex]))
							}

							worker.work.result.txsMutex.Unlock()
						}
					}
					listIndex += 1

					continue
				}

			}

		}

	}
}
