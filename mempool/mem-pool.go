package mempool

import (
	"bytes"
	"fmt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"sort"
	"strconv"
	"sync"
	"time"
)

type memPoolTx struct {
	tx          *transaction.Transaction
	added       int64
	mine        bool
	feePerByte  uint64
	feeToken    []byte
	chainHeight uint64

	sync.RWMutex `json:"-"`
}

type memPoolResult struct {
	txs          []*memPoolTx
	chainHash    cryptography.Hash
	chainHeight  uint64
	sync.RWMutex `json:"-"`
}

type memPoolOutput struct {
	hash    cryptography.Hash
	hashStr string
	tx      *memPoolTx `json:"-"`
}

type MemPool struct {
	txs      sync.Map
	txsCount uint64

	updateTask memPoolUpdateTask
	result     memPoolResult

	lockWritingTxs sync.RWMutex `json:"-"`
}

func (mempool *MemPool) AddTxToMemPoolSilent(tx *transaction.Transaction, height uint64, mine bool) (result bool, err error) {
	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()
	result = mempool.AddTxToMemPool(tx, height, mine)
	return
}

func (mempool *MemPool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) bool {

	//making sure that the transaction is not inserted twice
	mempool.lockWritingTxs.Lock()
	defer mempool.lockWritingTxs.Unlock()

	hash := tx.ComputeHash()
	if _, found := mempool.txs.Load(hash); found {
		return false
	}

	mempool.txsCount += 1
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
		nativeFee := config.NATIVE_TOKEN_STRING
		selectedFeeToken = &nativeFee
		selectedFee = fees.FEES_PER_BYTE[config.NATIVE_TOKEN_STRING]
	}

	if selectedFeeToken == nil {
		panic("Transaction fee was not accepted")
	}

	object := memPoolTx{
		tx:          tx,
		added:       time.Now().Unix(),
		feePerByte:  selectedFee / size,
		feeToken:    []byte(*selectedFeeToken),
		mine:        mine,
		chainHeight: height,
	}

	mempool.txs.Store(hash, &object)

	return true
}

func (mempool *MemPool) Exists(txId cryptography.Hash) bool {
	if _, exists := mempool.txs.Load(txId); exists {
		return true
	}
	return false
}

func (mempool *MemPool) Delete(txId cryptography.Hash) (tx *transaction.Transaction) {

	var exists bool
	if _, exists = mempool.txs.Load(txId); exists == false {
		return
	}

	mempool.lockWritingTxs.Lock()
	defer mempool.lockWritingTxs.Unlock()

	mempool.txs.Delete(txId)

	mempool.txsCount -= 1
	gui.Info2Update("mempool", strconv.FormatUint(mempool.txsCount, 10))

	return
}

func (mempool *MemPool) GetTxsList() []*transaction.Transaction {

	list := make([]*transaction.Transaction, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		tx := value.(*memPoolTx)
		list = append(list, tx.tx)
		return true
	})

	return list
}

func (mempool *MemPool) GetTxsListKeyValue() []*memPoolOutput {

	list := make([]*memPoolOutput, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(cryptography.Hash)
		tx := value.(*memPoolTx)
		list = append(list, &memPoolOutput{
			hash:    hash,
			hashStr: string(hash[:]),
			tx:      tx,
		})
		return true
	})

	return list
}

func (mempool *MemPool) GetTxsListKeyValueFilter(filter map[string]bool) []*memPoolOutput {

	list := make([]*memPoolOutput, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(cryptography.Hash)
		if !filter[string(hash[:])] {
			tx := value.(*memPoolTx)
			list = append(list, &memPoolOutput{
				hash:    hash,
				hashStr: string(hash[:]),
				tx:      tx,
			})
		}
		return true
	})

	return list
}

func (mempool *MemPool) Print() {

	mempool.lockWritingTxs.RLock()
	txsCount := mempool.txsCount
	mempool.lockWritingTxs.RUnlock()

	if txsCount == 0 {
		return
	}

	list := mempool.GetTxsListKeyValue()

	gui.Log("")
	for _, out := range list {
		out.tx.RLock()
		gui.Log(fmt.Sprintf("%20s %7d B %5d %32s", time.Unix(out.tx.added, 0).UTC().Format(time.RFC3339), len(out.tx.tx.Serialize()), out.tx.chainHeight, out.hash.String()))
		out.tx.RUnlock()
	}
	gui.Log("")

}

func (mempool *MemPool) UpdateChanges(hash cryptography.Hash, height uint64) {

	mempool.updateTask.Lock()
	defer mempool.updateTask.Unlock()

	mempool.updateTask.chainHash = hash
	mempool.updateTask.chainHeight = height
	mempool.updateTask.status = 1

}

func (mempool *MemPool) Refresh() {

	updateTask := memPoolUpdateTask{}
	hasWorkToDo := false

	var txList []*memPoolOutput
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
			mempool.result.txs = make([]*memPoolTx, 0)
			mempool.result.Unlock()
		}
		mempool.updateTask.RUnlock()

		if hasWorkToDo {

			if listIndex == -1 {

				txList = mempool.GetTxsListKeyValueFilter(txMap)
				if len(txList) > 0 {
					sort.Slice(txList, func(i, j int) bool {

						if txList[i].tx.feePerByte == txList[j].tx.feePerByte && txList[i].tx.tx.TxType == transaction_type.TxSimple && txList[j].tx.tx.TxType == transaction_type.TxSimple {
							return txList[i].tx.tx.TxBase.(*transaction_simple.TransactionSimple).Nonce < txList[j].tx.tx.TxBase.(*transaction_simple.TransactionSimple).Nonce
						}

						return txList[i].tx.feePerByte < txList[j].tx.feePerByte
					})

					var err error
					updateTask.boltTx, err = store.StoreBlockchain.DB.Begin(false)
					if err != nil {
						updateTask.CloseDB()
						time.Sleep(1000 * time.Millisecond)
						continue
					}
					reader := updateTask.boltTx.Bucket([]byte("Chain"))
					if !bytes.Equal(reader.Get([]byte("chainHash")), updateTask.chainHash[:]) {
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
								mempool.result.txs = append(mempool.result.txs, txList[listIndex].tx)
								mempool.result.Unlock()
								txList[listIndex].tx.Lock()
								txList[listIndex].tx.chainHeight = updateTask.chainHeight
								txList[listIndex].tx.Unlock()
							}
							listIndex += 1
						}()

						txMap[txList[listIndex].hashStr] = true
						txList[listIndex].tx.tx.IncludeTransaction(updateTask.chainHeight, updateTask.accs, updateTask.toks)
					}()

					continue
				}

			}

		} else {
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func (mempool *MemPool) GetNonce(publicKeyHash [20]byte) (result bool, nonce uint64) {

	txs := mempool.GetTxsList()
	for _, tx := range txs {
		if tx.TxType == transaction_type.TxSimple {
			base := tx.TxBase.(*transaction_simple.TransactionSimple)
			txPublicKeyHash := base.Vin[0].GetPublicKeyHash()
			if bytes.Equal(txPublicKeyHash[:], publicKeyHash[:]) {
				result = true
				if nonce <= base.Nonce {
					nonce = base.Nonce + 1
				}
			}
		}
	}

	return
}

func (mempool *MemPool) GetTransactions(blockHeight uint64, chainHash cryptography.Hash) []*transaction.Transaction {

	out := make([]*transaction.Transaction, 0)

	mempool.result.RLock()
	for _, mempoolTx := range mempool.result.txs {
		if bytes.Equal(mempool.result.chainHash[:], chainHash[:]) {
			out = append(out, mempoolTx.tx)
		}
	}
	mempool.result.RUnlock()

	return out
}

func InitMemPool() (mempool *MemPool) {

	gui.Log("MemPool init...")

	mempool = &MemPool{}

	go func() {
		for {
			time.Sleep(60 * time.Second)
			mempool.Print()
		}
	}()

	go mempool.Refresh()

	return
}
