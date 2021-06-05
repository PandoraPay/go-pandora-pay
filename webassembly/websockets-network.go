package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/accounts/account"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	api_common "pandora-pay/network/api/api-common"
	"syscall/js"
)

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
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
		socket := app.Network.Websockets.GetFirstSocket()
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

		data := socket.SendJSONAwaitAnswer([]byte("block-complete"), &api_common.APIBlockCompleteRequest{height, hash, api_common.RETURN_SERIALIZED})
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
		socket := app.Network.Websockets.GetFirstSocket()
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

		data := socket.SendJSONAwaitAnswer([]byte("tx"), &api_common.APIBlockCompleteRequest{height, hash, api_common.RETURN_SERIALIZED})
		if data.Err != nil {
			return nil, data.Err
		}

		received := &api_common.APITransactionSerialized{}
		if err = json.Unmarshal(data.Out, received); err != nil {
			return
		}

		final := &api_common.APITransaction{
			Tx:      &transaction.Transaction{},
			Mempool: received.Mempool,
		}

		if err = final.Tx.Deserialize(helpers.NewBufferReader(received.Tx)); err != nil {
			return
		}
		if err = final.Tx.BloomAll(); err != nil {
			return
		}
		return convertJSON(final)
	})
}

func subscribeNetworkAccount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var hash []byte
		if len(args) == 2 && len(args[1].String()) == 2*cryptography.PublicKeyHashHashSize {
			if hash, err = hex.DecodeString(args[1].String()); err != nil {
				return
			}
		}

		data := socket.SendJSONAwaitAnswer([]byte("sub/account"), &api_common.APIAccountRequest{args[0].String(), hash, api_common.RETURN_SERIALIZED})
		if data.Err != nil {
			return nil, data.Err
		}

		if data.Out == nil {
			return
		}

		acc := &account.Account{}
		if err = acc.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return
		}

		return convertJSON(acc)
	})
}

func subscribeNetworkAccount(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (out interface{}, err error) {

		go func() {

		}()

		return
	})
}
