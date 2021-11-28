package node_http_rpc

import (
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net/http"
	"pandora-pay/network/api/api_common"
)

type HTTPServerRPCRoutes struct {
	apiCommon *api_common.APICommon
}

func InitializeRPC(apiCommon *api_common.APICommon) (err error) {

	s := rpc.NewServer()

	s.RegisterCodec(json.NewCodec(), "application/json")
	if err = s.RegisterService(apiCommon, "api"); err != nil {
		return
	}

	http.Handle("/rpc/api/v1", s)

	return
}
