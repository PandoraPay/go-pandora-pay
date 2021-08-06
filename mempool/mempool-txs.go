package mempool

import (
	"encoding/hex"
	"fmt"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/recovery"
	"sync"
	"time"
)

type MempoolTxs struct {
	txsMap *sync.Map //[string]*mempoolTx
}

func (self *MempoolTxs) InsertTx(hashStr string, tx *mempoolTx) bool {
	_, stored := self.txsMap.LoadOrStore(hashStr, tx)
	return stored
}

func (self *MempoolTxs) DeleteTx(hashStr string) {
	self.txsMap.Delete(hashStr)
}

func (self *MempoolTxs) GetTxsFromMap() (out map[string]*mempoolTx) {

	out = make(map[string]*mempoolTx)
	self.txsMap.Range(func(key, value interface{}) bool {
		out[key.(string)] = value.(*mempoolTx)
		return true
	})

	return
}

func (self *MempoolTxs) GetTxsList() []*mempoolTx {
	data := self.GetTxsFromMap()
	out := make([]*mempoolTx, len(data))

	c := 0
	for _, tx := range data {
		out[c] = tx
		c += 1
	}
	return out
}

func (self *MempoolTxs) Exists(txId string) bool {
	_, loaded := self.txsMap.Load(txId)
	return loaded

}

func (self *MempoolTxs) Get(txId string) *mempoolTx {
	value, loaded := self.txsMap.Load(txId)
	if !loaded {
		return nil
	}
	return value.(*mempoolTx)
}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		&sync.Map{},
	}

	if config.DEBUG {
		recovery.SafeGo(func() {
			for {
				transactions := txs.GetTxsFromMap()
				if len(transactions) != 0 {
					gui.GUI.Log("")
					for _, out := range transactions {
						gui.GUI.Log(fmt.Sprintf("%12s %7d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
					}
					gui.GUI.Log("")
				}
				time.Sleep(60 * time.Second)
			}
		})
	}

	recovery.SafeGo(func() {
		//last := int64(-1)
		for {

			//txsCount := txs.GetTxsFromMap()
			//
			//if len(txsCount) != last {
			//	gui.GUI.Info2Update("mempool", strconv.FormatInt(txsCount, 10))
			//	last = txsCount
			//}
			//
			//count := 0
			//txs.txs.Range(func(key, value interface{}) bool {
			//	count += 1
			//	return true
			//})
			//gui.GUI.Info2Update("mempool2", strconv.Itoa(count))

			time.Sleep(1 * time.Second)
		}
	})

	return
}
