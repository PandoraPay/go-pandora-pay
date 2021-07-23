package mempool

import (
	"sync/atomic"
)

//written only by the thread
type MempoolResult struct {
	txs         *atomic.Value //[]*mempoolTx use atomic. Append Only
	totalSize   uint64        // used atomic
	chainHash   []byte        // 32bytes   ready only
	chainHeight uint64        // 		   read only
}
