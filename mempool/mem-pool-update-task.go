package mempool

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"sync"
)

type memPoolUpdateTask struct {
	boltTx       *bbolt.Tx
	chainHash    []byte //32 byte
	chainHeight  uint64
	accs         *accounts.Accounts
	toks         *tokens.Tokens
	status       int
	sync.RWMutex `json:"-"`
}

func (mempoolUpdateTask *memPoolUpdateTask) CloseDB() {
	if mempoolUpdateTask.boltTx != nil {
		mempoolUpdateTask.boltTx.Rollback()
		mempoolUpdateTask.boltTx = nil
	}
}
