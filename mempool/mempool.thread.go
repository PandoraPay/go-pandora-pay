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
	Result chan<- bool
}

//process the worker for transactions to prepare the transactions to the forger
func (worker *mempoolWorker) processing(
	suspendProcessingCn <-chan struct{},
	continueProcessingCn <-chan *mempoolWork, //SAFE
	addTransactionCn <-chan *MempoolWorkerAddTx,
	addToListCn chan<- *mempoolTx,
	removedFromListCn chan<- *mempoolTx,
) {

	work := <-continueProcessingCn

	var txList []*mempoolTx
	txMap := make(map[string]bool)

	for {

		if len(txList) > 1 {
			sortTxs(txList)
		}
		listIndex := 0

		//let's check hf the work has been changed
		store.StoreBlockchain.DB.View(func(dbTx store_db_interface.StoreDBTransactionInterface) (err error) {

			accs := accounts.NewAccounts(dbTx)
			toks := tokens.NewTokens(dbTx)

			var tx *mempoolTx
			var newAddTx *MempoolWorkerAddTx

			for {
				select {
				case <-suspendProcessingCn:
					return nil
				default:

					if listIndex == len(txList) {
						select {
						case _, _ = <-suspendProcessingCn:
							return nil
						case newAddTx, _ = <-addTransactionCn:
							tx = newAddTx.Tx
							if txMap[tx.Tx.Bloom.HashStr] {
								continue
							}
						}
					} else {
						tx = txList[listIndex]
						listIndex += 1
						newAddTx = nil
					}

					if err = tx.Tx.IncludeTransaction(work.chainHeight, accs, toks); err != nil {

						accs.Rollback()
						toks.Rollback()

						if newAddTx != nil {
							newAddTx.Result <- false
						} else {
							//removing
							txList = append(txList[:listIndex-1], txList[listIndex:]...)
							listIndex--
							removedFromListCn <- tx
						}

					} else {

						if work.result.totalSize+tx.Tx.Bloom.Size < config.BLOCK_MAX_SIZE {

							work.result.totalSize += tx.Tx.Bloom.Size
							work.result.txs.Store(append(work.result.txs.Load().([]*mempoolTx), tx))

							accs.Commit()
							toks.Commit()
						}

						if newAddTx != nil {

							txList = append(txList, newAddTx.Tx)

							listIndex += 1
							newAddTx.Result <- true
							addToListCn <- newAddTx.Tx
						}

					}

				}
			}

		})

		select {
		case newWork, ok := <-continueProcessingCn:
			if !ok {
				return
			}
			work = newWork
		}

	}
}
