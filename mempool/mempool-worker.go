package mempool

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"sort"
	"sync/atomic"
	"time"
	"unsafe"
)

//reset the forger
func (mempool *Mempool) UpdateWork(hash []byte, height uint64) {
	updateTask := &mempoolWorkTask{
		chainHash:   hash,
		chainHeight: height,
	}
	atomic.StorePointer(&mempool.updateTask, unsafe.Pointer(updateTask))
}

//process the worker for transactions to prepare the transactions to the forger
func (mempool *Mempool) processing() {

	var updateTask *mempoolWorkTask
	var updateTaskPointer unsafe.Pointer = nil
	updateTaskChanged := false

	hasWorkToDo := false

	var txList []*mempoolTx
	var txMap map[string]bool

	listIndex := -1
	for {

		//let's check hf the work has been changed
		pointer := atomic.LoadPointer(&mempool.updateTask)
		if pointer != updateTaskPointer {
			updateTaskPointer = pointer
			updateTask = (*mempoolWorkTask)(pointer)
			updateTask.CloseDB()
			hasWorkToDo = true
			txMap = make(map[string]bool)
			listIndex = -1
			updateTaskChanged = true

			mempool.result.Lock()
			mempool.result.chainHash = updateTask.chainHash
			mempool.result.chainHeight = updateTask.chainHeight
			mempool.result.txs = []*transaction.Transaction{}
			mempool.result.totalSize = 0
			mempool.result.Unlock()

		}

		if hasWorkToDo {

			if listIndex == -1 {

				mempool.txs.Lock()
				if mempool.txs.txsInserted > 0 || updateTaskChanged {
					if updateTaskChanged {
						txList = make([]*mempoolTx, len(mempool.txs.txsList))
						copy(txList, mempool.txs.txsList)
					} else {
						txList = make([]*mempoolTx, 0)
						for _, mempoolTx := range mempool.txs.txsList {
							if !txMap[mempoolTx.HashStr] {
								txList = append(txList, mempoolTx)
							}
						}
					}
				}
				mempool.txs.Unlock()
				updateTaskChanged = false

				if len(txList) > 0 {
					sort.Slice(txList, func(i, j int) bool {

						if txList[i].FeePerByte == txList[j].FeePerByte && txList[i].Tx.TxType == transaction_type.TxSimple && txList[j].Tx.TxType == transaction_type.TxSimple {
							return txList[i].Tx.TxBase.(*transaction_simple.TransactionSimple).Nonce < txList[j].Tx.TxBase.(*transaction_simple.TransactionSimple).Nonce
						}

						return txList[i].FeePerByte < txList[j].FeePerByte
					})

					var err error
					updateTask.boltTx, err = store.StoreBlockchain.DB.Begin(false)
					if err != nil {
						updateTask.CloseDB()
						time.Sleep(1000 * time.Millisecond)
						continue
					}
					updateTask.accs = accounts.NewAccounts(updateTask.boltTx)
					updateTask.toks = tokens.NewTokens(updateTask.boltTx)
				}
				listIndex = 0

			} else {

				if listIndex == len(txList) {
					updateTask.CloseDB()
					listIndex = -1
					time.Sleep(1000 * time.Millisecond)
					continue
				} else {

					func() {
						defer func() {
							if err := helpers.ConvertRecoverError(recover()); err != nil {
								updateTask.accs.Rollback()
								updateTask.toks.Rollback()
							} else {
								mempool.result.Lock()
								if mempool.result.totalSize+txList[listIndex].Size < config.BLOCK_MAX_SIZE {
									mempool.result.txs = append(mempool.result.txs, txList[listIndex].Tx)
									mempool.result.totalSize += txList[listIndex].Size
								}
								mempool.result.Unlock()
							}
							listIndex += 1
						}()

						txMap[txList[listIndex].HashStr] = true
						txList[listIndex].Tx.IncludeTransaction(updateTask.chainHeight, updateTask.accs, updateTask.toks)
					}()

					continue
				}

			}

		} else {
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func initWorker(mempool *Mempool) {
	go mempool.processing()
}
