package webassembly

import (
	"errors"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
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
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), &api_websockets.APIBlockRequest{uint64(args[0].Int()), []byte(args[1].String())})
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}
