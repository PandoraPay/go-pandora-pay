package mempool_sync

import (
	"encoding/json"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

type MempoolSync struct {
}

func DownloadMempool(conn *connection.AdvancedConnection) (err error) {

	out := conn.SendAwaitAnswer([]byte("mem-pool"), nil)
	if out.Err != nil {
		return
	}

	txs := make([]helpers.HexBytes, 0)
	if err = json.Unmarshal(out.Out, &txs); err != nil {
		return
	}

	return
}

func CreateMempoolSync() *MempoolSync {
	return &MempoolSync{}
}
