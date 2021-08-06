package mempool

import (
	"encoding/hex"
	"fmt"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/recovery"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type MempoolTxs struct {
	count        int32
	txsMap       *sync.Map //[string]*mempoolTx
	txsMapByKeys *sync.Map //[string]*( sync.map[string]bool )
}

func (self *MempoolTxs) InsertTx(hashStr string, tx *mempoolTx) bool {
	_, stored := self.txsMap.LoadOrStore(hashStr, tx)
	if stored {
		atomic.AddInt32(&self.count, 1)

		if config.SEED_WALLET_NODES_INFO {
			keys, _ := tx.Tx.GetAllKeys()
			for key := range keys {
				foundMap, _ := self.txsMapByKeys.LoadOrStore(key, &sync.Map{})
				foundMap.(*sync.Map).Store(key, tx)
			}
		}
	}
	return stored
}

func (self *MempoolTxs) DeleteTx(hashStr string) bool {
	txData, deleted := self.txsMap.LoadAndDelete(hashStr)
	if deleted {
		atomic.AddInt32(&self.count, -1)

		if config.SEED_WALLET_NODES_INFO {
			tx := txData.(*mempoolTx)
			keys, _ := tx.Tx.GetAllKeys()
			for key := range keys {
				foundMap, _ := self.txsMapByKeys.LoadOrStore(key, &sync.Map{})
				foundMap.(*sync.Map).Delete(key)
			}
		}
	}
	return deleted
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

func (self *MempoolTxs) GetAccountTxs(publicKeyHash []byte) map[string]*mempoolTx {
	if config.SEED_WALLET_NODES_INFO {

		out := make(map[string]*mempoolTx)

		foundMapData, _ := self.txsMapByKeys.Load(string(publicKeyHash))
		foundMap := foundMapData.(*sync.Map)
		foundMap.Range(func(key, value interface{}) bool {
			out[key.(string)] = value.(*mempoolTx)
			return true
		})

		return out
	}
	return nil
}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		0,
		&sync.Map{},
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
		last := int32(-1)
		for {

			txsCount := atomic.LoadInt32(&txs.count)

			if txsCount != last {
				gui.GUI.Info2Update("mempool", strconv.FormatInt(int64(txsCount), 10))
				last = txsCount
			}

			time.Sleep(1 * time.Second)
		}
	})

	return
}
