package api_common

import (
	"net/http"
)

type APIPingReply struct {
	Ping string `json:"ping" msgpack:"ping"`
}

func (api *APICommon) GetPing(r *http.Request, args *struct{}, reply *APIPingReply) error {
	reply.Ping = "pong"
	return nil
}
