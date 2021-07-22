package mempool

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type mempoolWork struct {
	chainHash   []byte         `json:"-"` //32 byte
	chainHeight uint64         `json:"-"`
	result      *MempoolResult `json:"-"`
}

type mempoolWorker struct {
	dbTx store_db_interface.StoreDBTransactionInterface `json:"-"`
}

type MempoolWorkerAddTx struct {
	Tx     *mempoolTx
	Result chan<- error
}

//process the worker for transactions to prepare the transactions to the forger
func (worker *mempoolWorker) processing(
	newWorkCn <-chan *mempoolWork,
	suspendProcessingCn <-chan struct{},
	continueProcessingCn <-chan struct{},
	addTransactionCn <-chan *MempoolWorkerAddTx,
	txs *MempoolTxs,
) {

	var work *mempoolWork

	txList := []*mempoolTx{}
	listIndex := 0
	txMap := make(map[string]bool)
	suspended := false
	readyListSent := false

	txs.clearList()

	for {

		select {
		case newWork := <-newWorkCn:
			work = newWork
			listIndex = 0
			txMap = make(map[string]bool)
			readyListSent = false
			txs.clearList()
		case <-continueProcessingCn:
			suspended = false
		}

		if work == nil || suspended {
			continue
		}

		if len(txList) > 1 {
			sortTxs(txList)
		}

		//let's check hf the work has been changed
		store.StoreBlockchain.DB.View(func(dbTx store_db_interface.StoreDBTransactionInterface) (err error) {

			accs := accounts.NewAccounts(dbTx)
			toks := tokens.NewTokens(dbTx)

			for {
				select {
				case newWork := <-newWorkCn:
					work = newWork
					listIndex = 0
					txMap = make(map[string]bool)
					readyListSent = false
					txs.clearList()
				case <-suspendProcessingCn:
					suspended = true
					return
				default:

					var tx *mempoolTx
					var newAddTx *MempoolWorkerAddTx

					if listIndex == len(txList) {

						if !readyListSent {
							txs.readyList()
							readyListSent = true
						}

						select {
						case newWork := <-newWorkCn:
							work = newWork
							listIndex = 0
							txMap = make(map[string]bool)
							readyListSent = false
							txs.clearList()
						case <-suspendProcessingCn:
							suspended = true
							return
						case newAddTx = <-addTransactionCn:
							tx = newAddTx.Tx
						}
					} else {
						tx = txList[listIndex]
						listIndex += 1
						newAddTx = nil
					}

					var finalErr error

					if tx != nil && !txMap[tx.Tx.Bloom.HashStr] {

						txMap[tx.Tx.Bloom.HashStr] = true

						if err = tx.Tx.IncludeTransaction(work.chainHeight, accs, toks); err != nil {

							finalErr = err

							accs.Rollback()
							toks.Rollback()

							if newAddTx == nil {
								//removing
								//this is done because listIndex was incremented already before
								txList = append(txList[:listIndex-1], txList[listIndex:]...)
								listIndex--
								delete(txMap, tx.Tx.Bloom.HashStr)
							}

						} else {

							if work.result.totalSize+tx.Tx.Bloom.Size < config.BLOCK_MAX_SIZE {

								work.result.totalSize += tx.Tx.Bloom.Size
								work.result.txs.Store(append(work.result.txs.Load().([]*mempoolTx), tx))

								accs.CommitChanges()
								toks.CommitChanges()
							}

							if newAddTx != nil {
								txList = append(txList, newAddTx.Tx)
								listIndex += 1
							}

							txs.addToList(tx)

						}

					}

					if newAddTx != nil && newAddTx.Result != nil {
						newAddTx.Result <- finalErr
					}

				}
			}

		})

	}
}
