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

type MempoolTxsData struct {
	txsCount int64
	txsList  []*mempoolTx
}

type MempoolTxs struct {
	txs *sync.Map // [string]*transaction.Transaction

	data             *atomic.Value //*MempoolTxsData
	waitTxsListReady *atomic.Value //chan <- interface{}

	lock                        *sync.Mutex
	temporary                   *MempoolTxsData
	temporaryWaitTxsListReadyCn chan struct{}
	stored                      bool
}

func (self *MempoolTxs) GetTxsList() (out []*mempoolTx) {

	<-self.waitTxsListReady.Load().(chan struct{})

	return self.data.Load().(*MempoolTxsData).txsList
}

func (self *MempoolTxs) Exists(txId string) *transaction.Transaction {

	value, loaded := self.txs.Load(txId)
	if !loaded {
		return nil
	}
	return value.(*transaction.Transaction)

}

func (self *MempoolTxs) clearList() {

	self.lock.Lock()
	defer self.lock.Unlock()

	if self.stored {
		self.temporaryWaitTxsListReadyCn = make(chan struct{})
		self.waitTxsListReady.Store(self.temporaryWaitTxsListReadyCn)

		self.stored = false

		self.temporary = &MempoolTxsData{
			0,
			[]*mempoolTx{},
		}
	}

}

func (self *MempoolTxs) readyList() {

	self.lock.Lock()
	defer self.lock.Unlock()

	if !self.stored {
		self.data.Store(self.temporary)
		self.stored = true
		close(self.temporaryWaitTxsListReadyCn)
	}
}

func (self *MempoolTxs) addToList(tx *mempoolTx) {

	self.lock.Lock()
	defer self.lock.Unlock()

	if self.stored {

		self.temporary = &MempoolTxsData{
			self.temporary.txsCount + 1,
			append(self.temporary.txsList, tx),
		}
		self.data.Store(self.temporary)

	} else {
		self.temporary.txsCount += 1
		self.temporary.txsList = append(self.temporary.txsList, tx)
	}

}

func createMempoolTxs() (txs *MempoolTxs) {

	txs = &MempoolTxs{
		&sync.Map{},
		&atomic.Value{}, //interface{}
		&atomic.Value{}, //interface{}
		&sync.Mutex{},
		&MempoolTxsData{
			0,
			[]*mempoolTx{},
		},
		make(chan struct{}),
		false,
	}
	txs.data.Store(&MempoolTxsData{
		0,
		[]*mempoolTx{},
	})
	txs.waitTxsListReady.Store(txs.temporaryWaitTxsListReadyCn)

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

			<-txs.waitTxsListReady.Load().(chan struct{})
			txsCount := txs.data.Load().(*MempoolTxsData).txsCount

			if txsCount != last {
				gui.GUI.Info2Update("mempool", strconv.FormatInt(txsCount, 10))
				last = txsCount
			}

			count := 0
			txs.txs.Range(func(key, value interface{}) bool {
				count += 1
				return true
			})
			gui.GUI.Info2Update("mempool2", strconv.Itoa(count))

			time.Sleep(1 * time.Second)
		}
	})

	return
}
