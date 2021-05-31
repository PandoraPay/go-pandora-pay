package webassembly

import (
	"encoding/hex"
	"errors"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/config/globals"
	"pandora-pay/helpers"
	"pandora-pay/network"
	api_common "pandora-pay/network/api/api-common"
	"syscall/js"
)

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := globals.Data["network"].(*network.Network).Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		var hash []byte
		if len(args) >= 2 {
			if hash, err = hex.DecodeString(args[1].String()); err != nil {
				return
			}
		}
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), &api_common.APIBlockRequest{uint64(args[0].Int()), hash})
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}

func getNetworkBlockComplete(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := globals.Data["network"].(*network.Network).Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var height uint64
		var hash []byte

		switch args[0].Type() {
		case js.TypeNumber:
			height = uint64(args[0].Int())
		case js.TypeString:
			if hash, err = hex.DecodeString(args[0].String()); err != nil {
				return
			}
		default:
			return nil, errors.New("Invalid argument")
		}

		data := socket.SendJSONAwaitAnswer([]byte("block-complete"), &api_common.APIBlockCompleteRequest{height, hash, 1})
		if data.Err != nil {
			return nil, data.Err
		}
		blkComplete := block_complete.CreateEmptyBlockComplete()
		if err := blkComplete.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return nil, err
		}
		if err := blkComplete.BloomAll(); err != nil {
			return nil, err
		}
		return convertJSON(blkComplete)
	})
}

func getNetworkTransaction(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := globals.Data["network"].(*network.Network).Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var height uint64
		var hash []byte

		switch args[0].Type() {
		case js.TypeNumber:
			height = uint64(args[0].Int())
		case js.TypeString:
			hash, err = hex.DecodeString(args[0].String())
		default:
			err = errors.New("Invalid argument")
		}
		if err != nil {
			return
		}

		data := socket.SendJSONAwaitAnswer([]byte("tx"), &api_common.APIBlockCompleteRequest{height, hash, 1})
		if data.Err != nil {
			return nil, data.Err
		}
		blkComplete := block_complete.CreateEmptyBlockComplete()
		if err = blkComplete.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return
		}
		if err = blkComplete.BloomAll(); err != nil {
			return
		}
		return convertJSON(blkComplete)
	})
}
