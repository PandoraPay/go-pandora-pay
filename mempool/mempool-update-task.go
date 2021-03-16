package mempool

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"sync"
)

type mempoolWorkTask struct {
	boltTx      *bbolt.Tx
	chainHash   []byte //32 byte
	chainHeight uint64
	accs        *accounts.Accounts
	toks        *tokens.Tokens
	status      int
}

type mempoolResult struct {
	txs          []*transaction.Transaction
	totalSize    uint64
	chainHash    []byte //32
	chainHeight  uint64
	sync.RWMutex `json:"-"`
}

func (mempoolWork *mempoolWorkTask) CloseDB() {
	if mempoolWork.boltTx != nil {
		mempoolWork.boltTx.Rollback()
		mempoolWork.boltTx = nil
	}
}
