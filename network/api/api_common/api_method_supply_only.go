package api_common

import (
	"net/http"
)

func (api *APICommon) GetSupplyOnly(r *http.Request, args *struct{}, reply *uint64) error {
	x := api.localChain.Load()
	*reply = x.Supply
	return nil
}
