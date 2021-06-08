package mempool_sync

import (
	"encoding/json"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
)

type MempoolSync struct {
	websockets *websocks.Websockets
}

func (mempoolSync *MempoolSync) DownloadMempool(conn *connection.AdvancedConnection) (err error) {

	cb := mempoolSync.websockets.ApiWebsockets.GetMap["mem-pool/new-tx-id"]

	start, times := 0, 0

	//times is used to avoid infinite loops
	for {

		out := conn.SendJSONAwaitAnswer([]byte("mem-pool"), &api_types.APIMempoolRequest{start})
		if out.Err != nil {
			return
		}

		data := &api_types.APIMempoolAnswer{}
		if err = json.Unmarshal(out.Out, data); err != nil {
			return
		}

		for _, tx := range data.Hashes {
			cb(conn, tx)
		}

		start += len(data.Hashes)
		times++

		if start >= data.Count || times > 10 {
			break
		}

	}

	return
}

func CreateMempoolSync(websockets *websocks.Websockets) *MempoolSync {
	return &MempoolSync{
		websockets: websockets,
	}
}
