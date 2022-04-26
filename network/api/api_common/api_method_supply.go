package api_common

import (
	"net/http"
	"pandora-pay/config/config_coins"
)

type APISupply struct {
	Supply    uint64 `json:"supply" msgpack:"supply"`
	MaxSupply uint64 `json:"maxSupply" msgpack:"maxSupply"`
}

func (api *APICommon) GetSupply(r *http.Request, args *struct{}, reply *APISupply) error {
	x := api.localChain.Load()
	reply.Supply = x.Supply
	reply.MaxSupply = config_coins.MAX_SUPPLY_COINS_UNITS
	return nil
}
