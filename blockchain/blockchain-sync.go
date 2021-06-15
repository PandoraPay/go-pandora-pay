package blockchain

import (
	blockchain_types "pandora-pay/blockchain/blockchain-types"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"strconv"
	"sync/atomic"
	"time"
)

type BlockchainSync struct {
	syncData            *atomic.Value               //*blockchain_types.BlockchainSyncData
	UpdateSyncMulticast *multicast.MulticastChannel `json:"-"` //chan *blockchain_types.BlockchainSyncData
}

func (chainSync *BlockchainSync) GetSyncTime() uint64 {
	chainSyncData := chainSync.syncData.Load().(*blockchain_types.BlockchainSyncData)
	return chainSyncData.SyncTime
}

func (chainSync *BlockchainSync) addBlocksChanged(blocks uint32, propagateNotification bool) *blockchain_types.BlockchainSyncData {

	chainSyncData := chainSync.syncData.Load().(*blockchain_types.BlockchainSyncData)

	newChainSyncData := &blockchain_types.BlockchainSyncData{
		BlocksChangedLastInterval: chainSyncData.BlocksChangedLastInterval + blocks,
	}

	if newChainSyncData.BlocksChangedLastInterval < 3 {
		newChainSyncData.Sync = chainSyncData.Sync
		newChainSyncData.SyncTime = chainSyncData.SyncTime
	}

	if propagateNotification {
		chainSync.UpdateSyncMulticast.BroadcastAwait(newChainSyncData)
	}

	return newChainSyncData
}

func (chainSync *BlockchainSync) resetBlocksChanged(propagateNotification bool) *blockchain_types.BlockchainSyncData {

	chainSyncData := chainSync.syncData.Load().(*blockchain_types.BlockchainSyncData)

	newChainSyncData := &blockchain_types.BlockchainSyncData{}

	if chainSyncData.BlocksChangedLastInterval < 3 {
		newChainSyncData.SyncTime = uint64(time.Now().Unix())
		newChainSyncData.Sync = true
	}

	if propagateNotification {
		chainSync.UpdateSyncMulticast.BroadcastAwait(newChainSyncData)
	}

	return newChainSyncData
}

func (chainSync *BlockchainSync) start() {

	recovery.SafeGo(func() {
		for {

			time.Sleep(time.Minute)
			chainSync.resetBlocksChanged(true)

		}
	})

	recovery.SafeGo(func() {
		for {
			chainSyncData := chainSync.syncData.Load().(*blockchain_types.BlockchainSyncData)

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

func createBlockchainSync() (out *BlockchainSync) {

	out = &BlockchainSync{
		syncData:            &atomic.Value{},
		UpdateSyncMulticast: multicast.NewMulticastChannel(),
	}
	out.syncData.Store(&blockchain_types.BlockchainSyncData{})

	out.start()

	return
}
