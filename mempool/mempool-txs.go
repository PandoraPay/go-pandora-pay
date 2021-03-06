package mempool

import (
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/recovery"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type MempoolTxs struct {
	txsMap           *sync.Map     //*mempoolTx
	txsCount         int64         //use atomic
	txsList          *atomic.Value //[]*mempoolTx
	addToListCn      chan *mempoolTx
	removeFromListCn chan *mempoolTx
	clearListCn      chan interface{}
}

func (self *MempoolTxs) GetTxsList() (out []*mempoolTx) {
	return self.txsList.Load().([]*mempoolTx)
}

func (self *MempoolTxs) Exists(txId string) *transaction.Transaction {
	out, _ := self.txsMap.Load(txId)
	if out == nil {
		return nil
	}
	return out.(*mempoolTx).Tx
}

func (self *MempoolTxs) process() {
	for {
		select {

		case <-self.clearListCn:
			list := self.txsList.Load().([]*mempoolTx)
			self.txsList.Store([]*mempoolTx{})
			atomic.StoreInt64(&self.txsCount, 0)
			for _, v := range list {
				self.txsMap.Delete(v.Tx.Bloom.HashStr)
			}
		case tx := <-self.addToListCn:
			self.txsList.Store(append(self.txsList.Load().([]*mempoolTx), tx))
			atomic.AddInt64(&self.txsCount, 1)
			self.txsMap.Store(tx.Tx.Bloom.HashStr, tx)
		case tx := <-self.removeFromListCn:
			list := self.txsList.Load().([]*mempoolTx)
			for i, tx2 := range list {
				if tx2 == tx {

					//removing atomic.Value array
					list2 := make([]*mempoolTx, len(list)-1)
					copy(list2, list)
					if len(list) > 1 && i != len(list)-1 {
						list2[i] = list[len(list)-1]
					}

					self.txsList.Store(list2)
					atomic.AddInt64(&self.txsCount, -1)

					self.txsMap.Delete(tx.Tx.Bloom.HashStr)
					break
				}
			}
		}
	}
}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		&sync.Map{},
		0,
		&atomic.Value{}, //[]*mempoolTx
		make(chan *mempoolTx),
		make(chan *mempoolTx),
		make(chan interface{}),
	}
	txs.txsList.Store([]*mempoolTx{})

	recovery.SafeGo(txs.process)

	if config.DEBUG {
		recovery.SafeGo(func() {
			for {
				transactions := txs.GetTxsList()
				if len(transactions) == 0 {
					return
				}

				gui.GUI.Log("")
				for _, out := range transactions {
					gui.GUI.Log(fmt.Sprintf("%12s %7d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
				}
				gui.GUI.Log("")
				time.Sleep(60 * time.Second)
			}
		})
	}

	recovery.SafeGo(func() {
		last := int64(-1)
		for {
			txsCount := atomic.LoadInt64(&txs.txsCount)
			if txsCount != last {
				gui.GUI.Info2Update("mempool", strconv.FormatInt(txsCount, 10))
				txsCount = last
			}
			time.Sleep(1 * time.Second)
		}
	})

	return
}
