package blockchain_sync

import (
	"fmt"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"time"
)

type BlockchainSyncData struct {
	SyncTime                      uint64 `json:"syncTime" msgpack:"syncTime" `
	BlocksChangedLastInterval     uint32 `json:"blocksChangedLastInterval" msgpack:"blocksChangedLastInterval"`
	BlocksChangedPreviousInterval uint32 `json:"blocksChangedPreviousInterval" msgpack:"blocksChangedPreviousInterval"`
	Sync                          bool   `json:"sync" msgpack:"sync" `
	Started                       bool   `json:"started" msgpack:"started" `
}

type BlockchainSync struct {
	syncData            *generics.Value[*BlockchainSyncData]
	UpdateSyncMulticast *multicast.MulticastChannel[*BlockchainSyncData]
	updateCn            chan *BlockchainSyncData
}

func (self *BlockchainSync) GetSyncData() *BlockchainSyncData {
	return self.syncData.Load()
}

func (self *BlockchainSync) GetSyncTime() uint64 {
	return self.syncData.Load().SyncTime
}

func (self *BlockchainSync) AddBlocksChanged(blocks uint32, propagateNotification bool) *BlockchainSyncData {

	chainSyncData := self.syncData.Load()

	newChainSyncData := &BlockchainSyncData{
		BlocksChangedPreviousInterval: chainSyncData.BlocksChangedPreviousInterval,
		BlocksChangedLastInterval:     chainSyncData.BlocksChangedLastInterval + blocks,
		Started:                       chainSyncData.Started,
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

	chainSyncData := self.syncData.Load()

	newChainSyncData := &BlockchainSyncData{
		BlocksChangedPreviousInterval: chainSyncData.BlocksChangedLastInterval,
		Started:                       chainSyncData.Started,
	}

	if chainSyncData.BlocksChangedLastInterval < 5 && (chainSyncData.Started || chainSyncData.BlocksChangedPreviousInterval < 4) {
		newChainSyncData.SyncTime = uint64(time.Now().Unix())
		newChainSyncData.Sync = true
		newChainSyncData.Started = true
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
			time.Sleep(2 * time.Minute)
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

func CreateBlockchainSync() (sync *BlockchainSync) {

	sync = &BlockchainSync{
		syncData:            &generics.Value[*BlockchainSyncData]{},
		UpdateSyncMulticast: multicast.NewMulticastChannel[*BlockchainSyncData](),
		updateCn:            make(chan *BlockchainSyncData),
	}

	if globals.Arguments["--skip-init-sync"] == true {
		sync.syncData.Store(&BlockchainSyncData{
			BlocksChangedPreviousInterval: 0,
			Started:                       true,
		})
		go func() {
			time.Sleep(1000 * time.Millisecond)
			sync.resetBlocksChanged(true)
		}()
	} else {
		sync.syncData.Store(&BlockchainSyncData{
			BlocksChangedPreviousInterval: 1000,
		})
	}

	sync.start()

	return
}
