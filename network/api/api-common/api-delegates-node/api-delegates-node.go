package api_delegates_node

import (
	"encoding/json"
	"pandora-pay/config"
	"pandora-pay/wallet"
)

type APIDelegatesNode struct {
	wallet *wallet.Wallet
}

func (api *APIDelegatesNode) getDelegatesInfo(request *ApiDelegatesNodeInfoRequest) ([]byte, error) {

	answer := &ApiDelegatesNodeInfoAnswer{
		config.DELEGATES_MAXIMUM,
		api.wallet.GetDelegatesCount(),
	}

	return json.Marshal(answer)
}

func CreateDelegatesNode(wallet *wallet.Wallet) (delegates *APIDelegatesNode) {

	delegates = &APIDelegatesNode{
		wallet,
	}

	return
}
