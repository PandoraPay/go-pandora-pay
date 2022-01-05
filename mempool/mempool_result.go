package mempool

import (
	"pandora-pay/helpers/generics"
)

//written only by the thread
type MempoolResult struct {
	txs         *generics.Value[[]*mempoolTx] // Append Only
	totalSize   uint64                        // used atomic
	chainHash   []byte                        // 32bytes   ready only
	chainHeight uint64                        // 		   read only
}
