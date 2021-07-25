package mempool

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"sync/atomic"
)

type mempoolWork struct {
	chainHash    []byte         `json:"-"` //32 byte
	chainHeight  uint64         `json:"-"`
	result       *MempoolResult `json:"-"`
	waitAnswerCn chan struct{}
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
	continueProcessingCn <-chan bool,
	addTransactionCn <-chan *MempoolWorkerAddTx,
	txs *MempoolTxs,
) {

	var work *mempoolWork

	txList := []*mempoolTx{}
	listIndex := 0
	txMap := make(map[string]bool)
	suspended := false
	notAllowedToContinue := false
	readyListSent := false

	var accs *accounts.Accounts
	var toks *tokens.Tokens

	includedTotalSize := uint64(0)
	includedTxs := []*mempoolTx{}

	txs.clearList()

	resetNow := func(newWork *mempoolWork) {

		if newWork.chainHash != nil {
			txs.clearList()
			readyListSent = false
		}
		close(newWork.waitAnswerCn)

		suspended = false

		if newWork.chainHash != nil {
			accs = nil
			toks = nil

			work = newWork
			includedTotalSize = uint64(0)
			includedTxs = []*mempoolTx{}
			listIndex = 0
			txMap = make(map[string]bool)
			if len(txList) > 1 {
				sortTxs(txList)
			}
		}
	}

	suspendNow := func() {
		suspended = true
		notAllowedToContinue = true
	}

	for {

		select {
		case newWork := <-newWorkCn:
			resetNow(newWork)
		case <-suspendProcessingCn:
			suspendNow()
		case suspend := <-continueProcessingCn:
			if suspend {
				suspended = true
			}
			notAllowedToContinue = false
		}

		if work == nil || suspended || notAllowedToContinue {
			continue
		}

		//let's check hf the work has been changed
		store.StoreBlockchain.DB.View(func(dbTx store_db_interface.StoreDBTransactionInterface) (err error) {

			if accs == nil {
				accs = accounts.NewAccounts(dbTx)
				toks = tokens.NewTokens(dbTx)
			} else {
				accs.Tx = dbTx
				toks.Tx = dbTx
			}

			for {
				select {
				case newWork := <-newWorkCn:
					resetNow(newWork)
				case <-suspendProcessingCn:
					suspendNow()
					return
				default:

					var tx *mempoolTx
					var newAddTx *MempoolWorkerAddTx

					if listIndex == len(txList) {

						//sending readyList only in case there is no transaction in the add channel
						if !readyListSent {

							select {
							case newAddTx = <-addTransactionCn:
								tx = newAddTx.Tx
							default:
								txs.readyList()
								readyListSent = true
							}

						}

						if tx == nil {
							select {
							case newWork := <-newWorkCn:
								resetNow(newWork)
							case <-suspendProcessingCn:
								suspendNow()
								return
							case newAddTx = <-addTransactionCn:
								tx = newAddTx.Tx
							}
						}

					} else {
						tx = txList[listIndex]
						listIndex += 1
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

							txs.txs.Delete(tx.Tx.Bloom.HashStr)

						} else {

							if includedTotalSize+tx.Tx.Bloom.Size < config.BLOCK_MAX_SIZE {

								includedTotalSize += tx.Tx.Bloom.Size
								includedTxs = append(includedTxs, tx)

								atomic.StoreUint64(&work.result.totalSize, includedTotalSize)
								work.result.txs.Store(includedTxs)

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
