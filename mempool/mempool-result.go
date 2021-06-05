package mempool

import (
	"sync/atomic"
)

//written only by the thread
type MempoolResult struct {
	txs         *atomic.Value //[]*mempoolTx
	totalSize   uint64        //
	chainHash   []byte        //  32bytes
	chainHeight uint64        // readOnly
}
