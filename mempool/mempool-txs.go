package mempool

import (
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

type MempoolTxsData struct {
	txsCount int64
	txsList  []*mempoolTx
}

type MempoolTxs struct {
	data             *atomic.Value //*MempoolTxsData
	waitTxsListReady *atomic.Value //chan <- interface{}

	addToListCn chan *mempoolTx
	readyListCn chan interface{}
	clearListCn chan interface{}
}

func (self *MempoolTxs) GetTxsList() (out []*mempoolTx) {

	<-self.waitTxsListReady.Load().(chan interface{})

	return self.data.Load().(*MempoolTxsData).txsList
}

func (self *MempoolTxs) Exists(txId string) *transaction.Transaction {

	<-self.waitTxsListReady.Load().(chan interface{})

	txList := self.data.Load().(*MempoolTxsData).txsList
	for _, tx := range txList {
		if tx.Tx.Bloom.HashStr == txId {
			return tx.Tx
		}
	}
	return nil
}

func (self *MempoolTxs) process() {

	data := &MempoolTxsData{
		0,
		[]*mempoolTx{},
	}
	stored := false

	for {
		select {

		case <-self.clearListCn:

			if stored {
				self.waitTxsListReady.Store(make(chan interface{}))
				stored = false
			}

			data = &MempoolTxsData{
				0,
				[]*mempoolTx{},
			}

		case <-self.readyListCn:

			cn := self.waitTxsListReady.Load().(chan interface{})
			close(cn)

			self.data.Store(data)
			stored = true

		case tx := <-self.addToListCn:
			if stored {
				newData := &MempoolTxsData{
					data.txsCount + 1,
					append(data.txsList, tx),
				}
				data = newData
				self.data.Store(data)
			} else {
				data.txsCount += 1
				data.txsList = append(data.txsList, tx)
			}
		}
	}
}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		&atomic.Value{}, //interface{}
		&atomic.Value{}, //interface{}
		make(chan *mempoolTx),
		make(chan interface{}),
		make(chan interface{}),
	}
	txs.data.Store(&MempoolTxsData{
		0,
		[]*mempoolTx{},
	})
	txs.waitTxsListReady.Store(make(chan interface{}))

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

			<-txs.waitTxsListReady.Load().(chan interface{})
			txsCount := txs.data.Load().(*MempoolTxsData).txsCount

			if txsCount != last {
				gui.GUI.Info2Update("mempool", strconv.FormatInt(txsCount, 10))
				txsCount = last
			}
			time.Sleep(1 * time.Second)
		}
	})

	return
}
