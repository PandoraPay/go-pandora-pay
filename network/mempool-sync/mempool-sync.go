package mempool_sync

import (
	"encoding/json"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
)

type MempoolSync struct {
	websockets *websocks.Websockets
}

func (mempoolSync *MempoolSync) DownloadMempool(conn *connection.AdvancedConnection) (err error) {

	out := conn.SendAwaitAnswer([]byte("mem-pool"), nil)
	if out.Err != nil {
		return
	}

	txs := make([]helpers.HexBytes, 0)
	if err = json.Unmarshal(out.Out, &txs); err != nil {
		return
	}

	for _, tx := range txs {

		cb := mempoolSync.websockets.ApiWebsockets.GetMap["mem-pool/new-tx-id"]
		cb(conn, tx)

	}

	return
}

func CreateMempoolSync(websockets *websocks.Websockets) *MempoolSync {
	return &MempoolSync{
		websockets: websockets,
	}
}
