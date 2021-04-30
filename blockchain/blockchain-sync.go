package blockchain

import (
	"pandora-pay/gui"
	"strconv"
	"sync/atomic"
	"time"
)

type BlockchainSync struct {
	SyncTime                uint64 `json:"-"` // use atomic
	BlocksChangedLastMinute uint32 `json:"-"` //use atomic
}

func (chainSync *BlockchainSync) updateBlockchainSyncInfo() {
	syncTime := atomic.LoadUint64(&chainSync.SyncTime)
	if syncTime != 0 {
		gui.Info2Update("Sync", time.Unix(int64(syncTime), 0).Format("2006-01-02 15:04:05"))
	} else {
		gui.Info2Update("Sync", "FALSE")
	}
	gui.Info2Update("Sync Blocks", strconv.FormatUint(uint64(atomic.LoadUint32(&chainSync.BlocksChangedLastMinute)), 10))
}

func (chainSync *BlockchainSync) SetSyncValue(sync bool) {
	if sync {
		atomic.CompareAndSwapUint64(&chainSync.SyncTime, 0, uint64(time.Now().Unix()))
	} else {
		atomic.SwapUint64(&chainSync.SyncTime, 0)
	}
}

func (chainSync *BlockchainSync) addBlocksChanged(blocks uint32) {
	atomic.AddUint32(&chainSync.BlocksChangedLastMinute, blocks)
}

func (chainSync *BlockchainSync) start() {
	go func() {
		for {
			time.Sleep(time.Minute)

			blocksLastMinute := atomic.SwapUint32(&chainSync.BlocksChangedLastMinute, 0)
			if blocksLastMinute > 2 {
				chainSync.SetSyncValue(true)
			} else {
				chainSync.SetSyncValue(false)
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
		BlocksChangedLastMinute: 0,
	}

	out.start()

	return
}
