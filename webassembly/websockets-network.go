package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common/api_types"
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
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), &api_types.APIBlockRequest{uint64(args[0].Int()), hash})
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

		data := socket.SendJSONAwaitAnswer([]byte("block-complete"), &api_types.APIBlockCompleteRequest{height, hash, api_types.RETURN_SERIALIZED})
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

func getNetworkAccount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var hash []byte
		if hash, err = hex.DecodeString(args[0].String()); err != nil {
			return
		}

		data := socket.SendJSONAwaitAnswer([]byte("account"), &api_types.APIAccountRequest{"", hash, api_types.RETURN_SERIALIZED})
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		acc := &account.Account{}
		if err = acc.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return
		}

		return convertJSON(acc)
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

		data := socket.SendJSONAwaitAnswer([]byte("tx"), &api_types.APIBlockCompleteRequest{height, hash, api_types.RETURN_SERIALIZED})
		if data.Err != nil {
			return nil, data.Err
		}

		received := &api_types.APITransactionSerialized{}
		if err = json.Unmarshal(data.Out, received); err != nil {
			return
		}

		final := &api_types.APITransaction{
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

func getNetworkTokenInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		var hash []byte
		if hash, err = hex.DecodeString(args[0].String()); err != nil {
			return
		}
		data := socket.SendJSONAwaitAnswer([]byte("token-info"), &api_types.APITokenInfoRequest{hash})
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}

func getNetworkToken(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		var hash []byte
		if hash, err = hex.DecodeString(args[0].String()); err != nil {
			return
		}
		data := socket.SendJSONAwaitAnswer([]byte("token"), &api_types.APITokenRequest{hash, api_types.RETURN_SERIALIZED})
		if data.Err != nil {
			return nil, data.Err
		}
		tok := &token.Token{}
		if err = tok.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return
		}
		return convertJSON(tok)
	})
}

func subscribeNetwork(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var key []byte
		if key, err = hex.DecodeString(args[0].String()); err != nil {
			return
		}

		_ = &api_types.APISubscriptionRequest{key, api_types.SubscriptionType(args[1].Int()), api_types.RETURN_SERIALIZED}
		data := socket.SendJSONAwaitAnswer([]byte("sub"), &api_types.APISubscriptionRequest{key, api_types.SubscriptionType(args[1].Int()), api_types.RETURN_SERIALIZED})
		return true, data.Err
	})
}

func unsubscribeNetwork(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var key []byte
		if key, err = hex.DecodeString(args[0].String()); err != nil {
			return
		}

		data := socket.SendJSONAwaitAnswer([]byte("unsub"), &api_types.APIUnsubscription{key, api_types.SubscriptionType(args[1].Int())})
		return true, data.Err
	})
}
