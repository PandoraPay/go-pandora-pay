package mempool

import (
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type MempoolAccountTxs struct {
	txs     map[string]*mempoolTx
	deleted bool
	sync.RWMutex
}

type MempoolTxs struct {
	count                     int32
	txsMap                    *generics.Map[string, *mempoolTx]
	accountsMapTxs            *generics.Map[string, *MempoolAccountTxs]
	UpdateMempoolTransactions *multicast.MulticastChannel[*blockchain_types.MempoolTransactionUpdate]
}

func (self *MempoolTxs) insertTx(tx *mempoolTx) bool {
	_, loaded := self.txsMap.LoadOrStore(tx.Tx.Bloom.HashStr, tx)
	if !loaded {
		atomic.AddInt32(&self.count, 1)
	}
	return !loaded
}

func (self *MempoolTxs) inserted(tx *mempoolTx) {
	if config.SEED_WALLET_NODES_INFO {

		keys := tx.Tx.GetAllKeys()
		for key := range keys {

			for {
				foundMap, _ := self.accountsMapTxs.LoadOrStore(key, &MempoolAccountTxs{})

				foundMap.Lock()
				if foundMap.deleted {
					foundMap.Unlock()
					continue
				}
				if foundMap.txs == nil {
					foundMap.txs = make(map[string]*mempoolTx)
				}
				foundMap.txs[tx.Tx.Bloom.HashStr] = tx
				foundMap.Unlock()
				break
			}
		}

		self.UpdateMempoolTransactions.Broadcast(&blockchain_types.MempoolTransactionUpdate{
			true,
			tx.Tx,
			false,
			keys,
		})

	}
}

func (self *MempoolTxs) deleteTx(hashStr string) bool {
	_, deleted := self.txsMap.LoadAndDelete(hashStr)
	if deleted {
		atomic.AddInt32(&self.count, -1)
	}
	return deleted
}

func (self *MempoolTxs) deleted(tx *mempoolTx, broadcastNotifications, includedInBlockchainNotification bool) {
	if config.SEED_WALLET_NODES_INFO {

		keys := tx.Tx.GetAllKeys()
		for key := range keys {
			foundMap, _ := self.accountsMapTxs.LoadOrStore(key, &MempoolAccountTxs{})

			foundMap.Lock()
			delete(foundMap.txs, tx.Tx.Bloom.HashStr)
			if len(foundMap.txs) == 0 {
				self.accountsMapTxs.Delete(key)
				foundMap.txs = nil
				foundMap.deleted = true
			}
			foundMap.Unlock()
		}

		if broadcastNotifications {
			self.UpdateMempoolTransactions.Broadcast(&blockchain_types.MempoolTransactionUpdate{
				false,
				tx.Tx,
				includedInBlockchainNotification,
				keys,
			})
		}

	}
}

func (self *MempoolTxs) GetTxsFromMap() (out map[string]*mempoolTx) {

	out = make(map[string]*mempoolTx)
	self.txsMap.Range(func(key string, value *mempoolTx) bool {
		out[key] = value
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
	return value
}

func (self *MempoolTxs) GetAccountTxs(publicKey []byte) []*mempoolTx {
	if config.SEED_WALLET_NODES_INFO {
		if foundMap, found := self.accountsMapTxs.Load(string(publicKey)); found {

			foundMap.RLock()

			out := make([]*mempoolTx, len(foundMap.txs))

			c := 0
			for _, tx := range foundMap.txs {
				out[c] = tx
				c += 1
			}

			foundMap.RUnlock()

			return out
		}
	}
	return nil
}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		0,
		&generics.Map[string, *mempoolTx]{},
		&generics.Map[string, *MempoolAccountTxs]{},
		multicast.NewMulticastChannel[*blockchain_types.MempoolTransactionUpdate](),
	}

	//printing from time to time the mempool
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
