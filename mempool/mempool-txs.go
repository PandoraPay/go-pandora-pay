package mempool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/recovery"
	"strconv"
	"sync/atomic"
	"time"
)

type MempoolTxs struct {
	txsCount         int64
	txsList          *atomic.Value
	addToListCn      chan *mempoolTx
	removeFromListCn chan *mempoolTx
}

func (self *MempoolTxs) GetTxsList() (out []*mempoolTx) {
	return self.txsList.Load().([]*mempoolTx)
}

func (self *MempoolTxs) Exists(txId []byte) *transaction.Transaction {
	list := self.txsList.Load().([]*mempoolTx)
	for _, tx := range list {
		if bytes.Equal(tx.Tx.Bloom.Hash, txId) {
			return tx.Tx
		}
	}
	return nil
}

func (self *MempoolTxs) process() {
	for {
		select {
		case tx := <-self.addToListCn:
			self.txsList.Store(append(self.txsList.Load().([]*mempoolTx), tx))
			atomic.AddInt64(&self.txsCount, 1)
		case tx := <-self.removeFromListCn:
			list := self.txsList.Load().([]*mempoolTx)
			for i, tx2 := range list {
				if tx2 == tx {

					list2 := make([]*mempoolTx, len(list)-1)
					copy(list2, list)
					if len(list) > 1 && i != len(list)-1 {
						list2[i] = list[len(list)-1]
					}

					self.txsList.Store(list2)
					atomic.AddInt64(&self.txsCount, -1)
					break
				}
			}
		}
	}
}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		0,
		&atomic.Value{},
		make(chan *mempoolTx, 100),
		make(chan *mempoolTx, 100),
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
