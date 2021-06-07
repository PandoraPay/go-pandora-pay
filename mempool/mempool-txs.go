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
	list             *atomic.Value
	txsCount         int64
	addToListCn      chan *mempoolTx
	removeFromListCn chan *mempoolTx
}

func (self *MempoolTxs) GetTxsList() []*mempoolTx {
	return self.list.Load().([]*mempoolTx)
}

func (self *MempoolTxs) Exists(txId []byte) *transaction.Transaction {
	list := self.list.Load().([]*mempoolTx)
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
			self.list.Store(append(self.list.Load().([]*mempoolTx), tx))
			atomic.AddInt64(&self.txsCount, 1)
		case tx := <-self.removeFromListCn:
			list := self.list.Load().([]*mempoolTx)
			for i, tx2 := range list {
				if tx2 == tx {
					list[len(list)-1], list[i] = list[i], list[len(list)-1]
					list = list[:len(list)-1]
					self.list.Store(list)
					atomic.AddInt64(&self.txsCount, -1)
					break
				}
			}
		}
	}
}

func (self *MempoolTxs) print() {

	transactions := self.GetTxsList()
	if len(transactions) == 0 {
		return
	}

	gui.GUI.Log("")
	for _, out := range transactions {
		gui.GUI.Log(fmt.Sprintf("%12s %7d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC822), out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
	}
	gui.GUI.Log("")

}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		list:             &atomic.Value{},
		addToListCn:      make(chan *mempoolTx),
		removeFromListCn: make(chan *mempoolTx),
	}
	txs.list.Store([]*mempoolTx{})

	recovery.SafeGo(txs.process)

	if config.DEBUG {
		recovery.SafeGo(func() {
			for {
				txs.print()
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
