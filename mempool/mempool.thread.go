package mempool

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"sort"
	"sync/atomic"
	"time"
)

type mempoolWork struct {
	chainHash   []byte         `json:"-"` //32 byte
	chainHeight uint64         `json:"-"`
	result      *mempoolResult `json:"-"`
}

type mempoolWorker struct {
	work        *mempoolWork                                   `json:"-"`
	workChanged bool                                           `json:"-"`
	dbTx        store_db_interface.StoreDBTransactionInterface `json:"-"`
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

			listIndex = -1
			worker.work = work
			worker.workChanged = true
			txMap = make(map[string]bool)

		default:
		}

		if worker.work == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		txListAll := mempoolTxs.txsList.Load().([]*mempoolTx)
		if worker.workChanged { //it is faster to copy first
			txList = txListAll
			worker.workChanged = false
		} else {
			for _, tx := range txListAll {
				if !txMap[tx.Tx.Bloom.HashStr] {
					txList = append(txList, tx)
				}
			}
		}

		if len(txList) > 0 {
			sortTxs(txList)
		}

		if listIndex == len(txList)-1 {
			time.Sleep(1000 * time.Millisecond)
			continue
		} else {

			store.StoreBlockchain.DB.View(func(dbTx store_db_interface.StoreDBTransactionInterface) (err error) {

				accs := accounts.NewAccounts(dbTx)
				toks := tokens.NewTokens(dbTx)

				for _, tx := range txList {

					if txMap[tx.Tx.Bloom.HashStr] {
						continue
					}

					txMap[tx.Tx.Bloom.HashStr] = true
					if err := tx.Tx.IncludeTransaction(worker.work.chainHeight, accs, toks); err != nil {

						accs.Rollback()
						toks.Rollback()

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
							accs.Commit()
							toks.Commit()
						}

					}

				}

				return nil

			})

		}

	}
}
