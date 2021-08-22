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
	api_faucet "pandora-pay/network/api/api-common/api-faucet"
	"pandora-pay/network/api/api-common/api_types"
	"syscall/js"
)

func getNetworkFaucetCoins(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		data := socket.SendJSONAwaitAnswer([]byte("faucet/coins"), &api_faucet.APIFaucetCoinsRequest{args[0].String(), args[1].String()})
		if data.Err != nil {
			return nil, data.Err
		}
		return hex.EncodeToString(data.Out), nil
	})
}

func getNetworkFaucetInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		data := socket.SendJSONAwaitAnswer([]byte("faucet/info"), nil)
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		var hash []byte
		var err error
		if len(args) >= 2 {
			if hash, err = hex.DecodeString(args[1].String()); err != nil {
				return nil, err
			}
		}
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), &api_types.APIBlockInfoRequest{uint64(args[0].Int()), hash})
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}

func getNetworkBlockComplete(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var height uint64
		var hash []byte
		var err error

		switch args[0].Type() {
		case js.TypeNumber:
			height = uint64(args[0].Int())
		case js.TypeString:
			if hash, err = hex.DecodeString(args[0].String()); err != nil {
				return nil, err
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
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account"), &api_types.APIAccountRequest{api_types.APIAccountBaseRequest{"", hash}, api_types.RETURN_SERIALIZED})
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		acc := &account.Account{}
		if err = acc.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return nil, err
		}

		return convertJSON(acc)
	})
}

func getNetworkAccountTxs(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account/txs"), &api_types.APIAccountTxsRequest{api_types.APIAccountBaseRequest{"", hash}, uint64(args[1].Int())})
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		return string(data.Out), nil
	})
}

func getNetworkAccountMempool(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account/mem-pool"), &api_types.APIAccountBaseRequest{"", hash})
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		result := make([]helpers.HexBytes, 0)
		if err := json.Unmarshal(data.Out, &result); err != nil {
			return nil, err
		}

		return convertJSON(result)
	})
}

func getNetworkTransaction(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		var height uint64
		var hash []byte
		var err error

		switch args[0].Type() {
		case js.TypeNumber:
			height = uint64(args[0].Int())
		case js.TypeString:
			hash, err = hex.DecodeString(args[0].String())
		default:
			err = errors.New("Invalid argument")
		}
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("tx"), &api_types.APIBlockCompleteRequest{height, hash, api_types.RETURN_SERIALIZED})
		if data.Err != nil {
			return nil, data.Err
		}

		received := &api_types.APITransaction{}
		if err = json.Unmarshal(data.Out, received); err != nil {
			return nil, err
		}

		received.Tx = &transaction.Transaction{}
		if err = received.Tx.Deserialize(helpers.NewBufferReader(received.TxSerialized)); err != nil {
			return nil, err
		}
		if err = received.Tx.BloomAll(); err != nil {
			return nil, err
		}

		return convertJSON(received)
	})
}

func getNetworkTokenInfo(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}
		data := socket.SendJSONAwaitAnswer([]byte("token-info"), &api_types.APITokenInfoRequest{hash})
		if data.Err != nil {
			return nil, data.Err
		}
		return string(data.Out), nil
	})
}

func getNetworkToken(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}
		data := socket.SendJSONAwaitAnswer([]byte("token"), &api_types.APITokenRequest{hash, api_types.RETURN_SERIALIZED})
		if data.Err != nil {
			return nil, data.Err
		}
		tok := &token.Token{}
		if err = tok.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return nil, err
		}
		return convertJSON(tok)
	})
}

func getNetworkMempool(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		chainHash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("mem-pool"), &api_types.APIMempoolRequest{chainHash, args[1].Int(), args[2].Int()})
		if data.Err != nil {
			return nil, data.Err
		}

		return string(data.Out), nil
	})
}

func subscribeNetwork(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		req := &api_types.APISubscriptionRequest{key, api_types.SubscriptionType(args[1].Int()), api_types.RETURN_SERIALIZED}
		data := socket.SendJSONAwaitAnswer([]byte("sub"), req)
		return true, data.Err
	})
}

func unsubscribeNetwork(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("unsub"), &api_types.APIUnsubscriptionRequest{key, api_types.SubscriptionType(args[1].Int())})
		return true, data.Err
	})
}
