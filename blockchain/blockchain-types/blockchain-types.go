package blockchain_types

type BlockchainSyncData struct {
	SyncTime                  uint64 `json:"syncTime"`
	BlocksChangedLastInterval uint32 `json:"blocksChangedLastInterval"`
	Sync                      bool   `json:"sync"`
}
