package blockchain

import (
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
	"sync/atomic"
	"time"
)

type BlockchainSync struct {
	SyncTime                uint64                    `json:"-"` // use atomic
	blocksChangedLastMinute uint32                    `json:"-"` //use atomic
	UpdateSyncMulticast     *helpers.MulticastChannel `json:"-"` //chan uint64
}

func (chainSync *BlockchainSync) updateBlockchainSyncInfo() {
	syncTime := atomic.LoadUint64(&chainSync.SyncTime)
	if syncTime != 0 {
		gui.Info2Update("Sync", time.Unix(int64(syncTime), 0).Format("2006-01-02 15:04:05"))
	} else {
		gui.Info2Update("Sync", "FALSE")
	}
	gui.Info2Update("Sync Blocks", strconv.FormatUint(uint64(atomic.LoadUint32(&chainSync.blocksChangedLastMinute)), 10))
}

func (chainSync *BlockchainSync) SetSyncValue(sync bool) {

	if sync {
		newValue := uint64(time.Now().Unix())
		if atomic.CompareAndSwapUint64(&chainSync.SyncTime, 0, newValue) {
			chainSync.UpdateSyncMulticast.Broadcast(newValue)
		}
	} else {
		if atomic.LoadUint64(&chainSync.SyncTime) == 0 {
			atomic.SwapUint64(&chainSync.SyncTime, 0)
			chainSync.UpdateSyncMulticast.Broadcast(0)
		}
	}

}

func (chainSync *BlockchainSync) addBlocksChanged(blocks uint32) {
	blocksChangedLastMinute := atomic.AddUint32(&chainSync.blocksChangedLastMinute, blocks)
	if blocksChangedLastMinute > 2 {
		chainSync.SetSyncValue(false)
	}
}

func (chainSync *BlockchainSync) start() {
	go func() {
		for {
			time.Sleep(time.Minute)

			blocksChangedLastMinute := atomic.SwapUint32(&chainSync.blocksChangedLastMinute, 0)
			if blocksChangedLastMinute > 2 {
				chainSync.SetSyncValue(false)
			} else {
				chainSync.SetSyncValue(true)
			}

		}
	}()

	go func() {
		for {
			chainSync.updateBlockchainSyncInfo()
			time.Sleep(2 * time.Second)
		}
	}()
}

func createBlockchainSync() (out *BlockchainSync) {

	out = &BlockchainSync{
		SyncTime:                0,
		blocksChangedLastMinute: 0,
		UpdateSyncMulticast:     helpers.NewMulticastChannel(),
	}

	out.start()

	return
}
