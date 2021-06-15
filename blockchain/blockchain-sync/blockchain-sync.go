package blockchain_sync

import (
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"strconv"
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
		self.UpdateSyncMulticast.BroadcastAwait(newChainSyncData)
	}

	self.syncData.Store(newChainSyncData)

	return newChainSyncData
}

func (self *BlockchainSync) resetBlocksChanged(propagateNotification bool) *BlockchainSyncData {

	chainSyncData := self.syncData.Load().(*BlockchainSyncData)

	newChainSyncData := &BlockchainSyncData{}

	if chainSyncData.BlocksChangedLastInterval < 3 {
		newChainSyncData.SyncTime = uint64(time.Now().Unix())
		newChainSyncData.Sync = true
	}

	self.syncData.Store(newChainSyncData)

	if propagateNotification {
		self.UpdateSyncMulticast.BroadcastAwait(newChainSyncData)
	}

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
			chainSyncData := self.syncData.Load().(*BlockchainSyncData)

			if chainSyncData.SyncTime != 0 {
				gui.GUI.Info2Update("Sync", time.Unix(int64(chainSyncData.SyncTime), 0).Format("2006-01-02 15:04:05"))
			} else {
				gui.GUI.Info2Update("Sync", "FALSE")
			}
			gui.GUI.Info2Update("Sync Blocks", strconv.FormatUint(uint64(chainSyncData.BlocksChangedLastInterval), 10))
			time.Sleep(2 * time.Second)
		}
	})
}

func CreateBlockchainSync() (out *BlockchainSync) {

	out = &BlockchainSync{
		syncData:            &atomic.Value{},
		UpdateSyncMulticast: multicast.NewMulticastChannel(),
	}
	out.syncData.Store(&BlockchainSyncData{})

	out.start()

	return
}
