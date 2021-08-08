package mempool_sync

import (
	"encoding/json"
	"pandora-pay/config"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
)

type MempoolSync struct {
	websockets *websocks.Websockets
}

func (mempoolSync *MempoolSync) DownloadMempool(conn *connection.AdvancedConnection) (err error) {

	cb := mempoolSync.websockets.ApiWebsockets.GetMap["mem-pool/new-tx-id"]

	index, page := 0, 0
	count := config.API_MEMPOOL_MAX_TRANSACTIONS

	//times is used to avoid infinite loops
	for {

		out := conn.SendJSONAwaitAnswer([]byte("mem-pool"), &api_types.APIMempoolRequest{page, 0})
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

		index += len(data.Hashes)
		page++

		if len(data.Hashes) == 0 || len(data.Hashes) != count || index >= data.Count || page > 20 { //done
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
