package blockchain_sync

import (
	"fmt"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"sync/atomic"
	"time"
)

type BlockchainSyncData struct {
	SyncTime                  uint64 `json:"syncTime"`
	BlocksChangedLastInterval uint32 `json:"blocksChangedLastInterval"`
	Sync                      bool   `json:"sync"`
}

type BlockchainSync struct {
	syncData            *atomic.Value               //*BlockchainSyncData
	UpdateSyncMulticast *multicast.MulticastChannel `json:"-"` //chan *BlockchainSyncData
	updateCn            chan *BlockchainSyncData
}

func (self *BlockchainSync) GetSyncData() *BlockchainSyncData {
	return self.syncData.Load().(*BlockchainSyncData)
}

func (self *BlockchainSync) GetSyncTime() uint64 {
	return self.syncData.Load().(*BlockchainSyncData).SyncTime
}

func (self *BlockchainSync) AddBlocksChanged(blocks uint32, propagateNotification bool) *BlockchainSyncData {

	chainSyncData := self.syncData.Load().(*BlockchainSyncData)

	newChainSyncData := &BlockchainSyncData{
		BlocksChangedLastInterval: chainSyncData.BlocksChangedLastInterval + blocks,
	}

	if newChainSyncData.BlocksChangedLastInterval < 3 {
		newChainSyncData.Sync = chainSyncData.Sync
		newChainSyncData.SyncTime = chainSyncData.SyncTime
	}

	if propagateNotification {
		self.UpdateSyncMulticast.Broadcast(newChainSyncData)
	}

	self.syncData.Store(newChainSyncData)

	self.updateCn <- newChainSyncData

	return newChainSyncData
}

func (self *BlockchainSync) resetBlocksChanged(propagateNotification bool) *BlockchainSyncData {

	chainSyncData := self.syncData.Load().(*BlockchainSyncData)

	newChainSyncData := &BlockchainSyncData{}

	if chainSyncData.BlocksChangedLastInterval < 4 {
		newChainSyncData.SyncTime = uint64(time.Now().Unix())
		newChainSyncData.Sync = true
	}

	self.syncData.Store(newChainSyncData)

	if propagateNotification {
		self.UpdateSyncMulticast.Broadcast(newChainSyncData)
	}

	self.updateCn <- newChainSyncData

	return newChainSyncData
}

func (self *BlockchainSync) start() {

	recovery.SafeGo(func() {
		for {

			time.Sleep(time.Minute)
			self.resetBlocksChanged(true)

		}
	})

	recovery.SafeGo(func() {
		for {

			chainSyncData, ok := <-self.updateCn
			if !ok {
				return
			}

			if chainSyncData.SyncTime != 0 {
				gui.GUI.Info2Update("Sync", fmt.Sprintf("%s %d", time.Unix(int64(chainSyncData.SyncTime), 0).Format("15:04:05"), chainSyncData.BlocksChangedLastInterval))
			} else {
				gui.GUI.Info2Update("Sync", fmt.Sprintf("FALSE %d", chainSyncData.BlocksChangedLastInterval))
			}
		}
	})
}

func CreateBlockchainSync() (out *BlockchainSync) {

	out = &BlockchainSync{
		syncData:            &atomic.Value{},
		UpdateSyncMulticast: multicast.NewMulticastChannel(false),
		updateCn:            make(chan *BlockchainSyncData),
	}
	out.syncData.Store(&BlockchainSyncData{})

	out.start()

	return
}
