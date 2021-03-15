package mempool

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"sort"
	"strconv"
	"sync"
	"time"
)

type memPoolTx struct {
	Tx          *transaction.Transaction
	Added       int64
	Mine        bool
	Size        uint64
	FeePerByte  uint64
	FeeToken    []byte //20 byte
	ChainHeight uint64
	Hash        []byte //32 byte
	HashStr     string
}

type memPoolResult struct {
	txs          []*memPoolTx
	totalSize    uint64
	chainHash    []byte //32
	chainHeight  uint64
	sync.RWMutex `json:"-"`
}

type MemPool struct {
	txs            sync.Map
	txsCount       uint64
	txsList        []*memPoolTx
	lockWritingTxs sync.RWMutex `json:"-"`
	updateTask     *memPoolUpdateTask
	result         *memPoolResult
}

func (mempool *MemPool) AddTxToMemPoolSilent(tx *transaction.Transaction, height uint64, mine bool) (result bool, err error) {
	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()
	result = mempool.AddTxToMemPool(tx, height, mine)
	return
}

func (mempool *MemPool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) bool {

	hash := tx.ComputeHash()
	hashStr := string(hash)

	//making sure that the transaction is not inserted twice
	mempool.lockWritingTxs.Lock()
	defer mempool.lockWritingTxs.Unlock()

	if _, found := mempool.txs.Load(hashStr); found {
		return false
	}

	gui.Info2Update("mempool", strconv.FormatUint(mempool.txsCount, 10))

	minerFees := tx.ComputeFees()

	size := uint64(len(tx.Serialize()))
	var selectedFeeToken *string
	var selectedFee uint64

	for token := range fees.FEES_PER_BYTE {
		if minerFees[token] != 0 {
			feePerByte := minerFees[token] / size
			if feePerByte >= fees.FEES_PER_BYTE[token] {
				selectedFeeToken = &token
				selectedFee = minerFees[*selectedFeeToken]
				break
			}
		}
	}

	//if it is mine and no fee was paid, let's fake a fee
	if mine && selectedFeeToken == nil {
		selectedFeeToken = &config.NATIVE_TOKEN_STRING
		selectedFee = fees.FEES_PER_BYTE[config.NATIVE_TOKEN_STRING]
	}

	if selectedFeeToken == nil {
		panic("Transaction fee was not accepted")
	}

	object := memPoolTx{
		Tx:          tx,
		Added:       time.Now().Unix(),
		Size:        size,
		FeePerByte:  selectedFee / size,
		FeeToken:    []byte(*selectedFeeToken),
		Mine:        mine,
		ChainHeight: height,
		Hash:        hash,
		HashStr:     hashStr,
	}

	mempool.txsCount += 1
	mempool.txs.Store(hashStr, &object)
	mempool.txsList = append(mempool.txsList, &object)

	return true
}

func (mempool *MemPool) Exists(txId []byte) bool {
	if _, exists := mempool.txs.Load(string(txId)); exists {
		return true
	}
	return false
}

func (mempool *MemPool) Delete(txId []byte) (tx *transaction.Transaction) {

	hashStr := string(txId)
	if _, exists := mempool.txs.Load(hashStr); exists == false {
		return
	}

	mempool.lockWritingTxs.Lock()
	defer mempool.lockWritingTxs.Unlock()

	mempool.txs.Delete(hashStr)
	mempool.txsCount -= 1
	for i, txOut := range mempool.txsList {
		if txOut.HashStr == hashStr {
			mempool.txsList[len(mempool.txsList)-1], mempool.txsList[i] = mempool.txsList[i], mempool.txsList[len(mempool.txsList)-1]
			mempool.txsList = mempool.txsList[:len(mempool.txsList)-1]
			break
		}
	}

	gui.Info2Update("mempool", strconv.FormatUint(mempool.txsCount, 10))
	return
}

func (mempool *MemPool) UpdateChanges(hash []byte, height uint64) {

	mempool.updateTask.Lock()
	defer mempool.updateTask.Unlock()

	mempool.updateTask.chainHash = hash
	mempool.updateTask.chainHeight = height
	mempool.updateTask.status = 1

}

func (mempool *MemPool) Refresh() {

	updateTask := memPoolUpdateTask{}
	hasWorkToDo := false

	var txList []*memPoolTx
	var txMap map[string]bool

	listIndex := -1
	for {

		mempool.updateTask.RLock()
		if mempool.updateTask.status == 1 {

			updateTask.CloseDB()
			updateTask.chainHash = mempool.updateTask.chainHash
			updateTask.chainHeight = mempool.updateTask.chainHeight
			mempool.updateTask.status = 0
			hasWorkToDo = true

			txMap = make(map[string]bool)
			listIndex = -1

			mempool.result.Lock()
			mempool.result.chainHash = mempool.updateTask.chainHash
			mempool.result.chainHeight = mempool.updateTask.chainHeight
			mempool.result.txs = []*memPoolTx{}
			mempool.result.totalSize = 0
			mempool.result.Unlock()
		}
		mempool.updateTask.RUnlock()

		if hasWorkToDo {

			if listIndex == -1 {

				txList = mempool.GetTxsListKeyValueFilter(txMap)
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
									mempool.result.txs = append(mempool.result.txs, txList[listIndex])
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

func InitMemPool() (mempool *MemPool) {

	gui.Log("MemPool init...")

	mempool = &MemPool{
		updateTask: &memPoolUpdateTask{},
		result:     &memPoolResult{},
		txsList:    []*memPoolTx{},
	}

	go func() {
		for {
			time.Sleep(60 * time.Second)
			mempool.print()
		}
	}()

	go mempool.Refresh()

	return
}
