package webassembly

import (
	"errors"
	"pandora-pay/config/globals"
	"pandora-pay/network"
	api_websockets "pandora-pay/network/api/api-websockets"
	"syscall/js"
)

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (interface{}, error) {
		socket := globals.Data["network"].(*network.Network).Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), api_websockets.APIBlockHeight(args[0].Int()))
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}
