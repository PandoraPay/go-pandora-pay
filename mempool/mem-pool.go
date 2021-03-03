package mempool

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

var FeesPerByteAccepted map[string]uint64

type MemPoolTx struct {
	Tx         *transaction.Transaction
	FeePerByte uint64
	FeeToken   []byte
}

type MemPoolType struct {
	txs          sync.Map
	sortedByFees []helpers.Hash

	sync.Mutex
}

var MemPool MemPoolType

func (mempool *MemPoolType) AddTxToMemPool(tx *transaction.Transaction) (result bool) {

	var err error

	mempool.Lock()
	defer mempool.Unlock()

	hash := tx.ComputeHash()
	if _, found := mempool.txs.Load(hash); found {
		return
	}

	var fees map[string]uint64
	if fees, err = tx.ComputeFees(); err != nil {
		return false
	}
	size := uint64(len(tx.Serialize(true)))
	var selectedFeeToken string
	for token, feePerByteAccepted := range FeesPerByteAccepted {
		if fees[token] != 0 {
			feePerByte := fees[token] / size
			if feePerByte >= feePerByteAccepted {
				selectedFeeToken = token
				break
			}
		}
	}
	if selectedFeeToken == "" {
		return
	}

	object := MemPoolTx{
		Tx:         tx,
		FeePerByte: fees[selectedFeeToken] / size,
		FeeToken:   []byte(selectedFeeToken),
	}
	mempool.txs.Store(hash, object)

	return true
}

func InitMemPool() (err error) {

	gui.Info("MemPool init...")
	FeesPerByteAccepted = make(map[string]uint64)
	FeesPerByteAccepted[string(config.NATIVE_TOKEN)] = 10

	return
}
