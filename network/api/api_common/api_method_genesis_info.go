package api_common

import (
	"net/http"
	"pandora-pay/blockchain/genesis"
)

type APIGenesisInfoRequest struct {
}

type APIGenesisInfoReply struct {
	GenesisInfo *genesis.GenesisDataType `json:"genesisInfo" msgpack:"genesisInfo"`
}

func (api *APICommon) GetGenesisInfo(r *http.Request, args *APIGenesisInfoRequest, reply *APIGenesisInfoReply) error {

	reply.GenesisInfo = genesis.GenesisData

	return nil
}
