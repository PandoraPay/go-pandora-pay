package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/data_storage/accounts/account"
	plain_account "pandora-pay/blockchain/data_storage/plain-accounts/plain-account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/data_storage/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	api_faucet "pandora-pay/network/api/api-common/api-faucet"
	"pandora-pay/network/api/api-common/api_types"
	"strconv"
	"sync"
	"syscall/js"
	"time"
)

var txMutex sync.Mutex

func getNetworkFaucetCoins(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		data := socket.SendJSONAwaitAnswer([]byte("faucet/coins"), &api_faucet.APIFaucetCoinsRequest{args[0].String(), args[1].String()}, time.Minute)
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
		data := socket.SendJSONAwaitAnswer([]byte("faucet/info"), nil, 0)
		if data.Err != nil {
			return nil, data.Err
		}
		return convertBytes(data.Out)
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
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), &api_types.APIBlockInfoRequest{uint64(args[0].Int()), hash}, 0)
		if data.Err != nil {
			return nil, data.Err
		}

		return convertBytes(data.Out)
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

		data := socket.SendJSONAwaitAnswer([]byte("block-complete"), &api_types.APIBlockCompleteRequest{height, hash, api_types.RETURN_SERIALIZED}, 0)
		if data.Err != nil {
			return nil, data.Err
		}
		blkComplete := block_complete.CreateEmptyBlockComplete()

		if err := blkComplete.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return nil, err
		}
		for _, tx := range blkComplete.Txs {
			if err = tx.BloomExtraVerified(); err != nil {
				return nil, err
			}
		}

		return convertJSONBytes(blkComplete)
	})
}

func getNetworkAccountsCount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		token, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendAwaitAnswer([]byte("accounts/count"), token, 0)
		if data.Err != nil {
			return nil, data.Err
		}

		return strconv.ParseUint(string(data.Out), 10, 64)
	})
}

func getNetworkAccountsKeysByIndex(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		request := &api_types.APIAccountsKeysByIndexRequest{nil, nil, false}
		if err := unmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("accounts/keys-by-index"), request, 0)
		if data.Err != nil {
			return nil, data.Err
		}

		return convertBytes(data.Out)
	})
}

func getNetworkAccountsByKeys(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		request := &api_types.APIAccountsByKeysRequest{nil, nil, nil, false, api_types.RETURN_SERIALIZED}
		if err := unmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		request.ReturnType = api_types.RETURN_SERIALIZED

		data := socket.SendJSONAwaitAnswer([]byte("accounts/by-keys"), request, 0)
		if data.Err != nil {
			return nil, data.Err
		}

		return convertBytes(data.Out)
	})
}

func getNetworkAccount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account"), &api_types.APIAccountRequest{api_types.APIAccountBaseRequest{"", publicKey}, api_types.RETURN_SERIALIZED}, 0)
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		result := &api_types.APIAccount{}
		if err = json.Unmarshal(data.Out, result); err != nil {
			return nil, err
		}

		if result != nil {

			result.Accs = make([]*account.Account, len(result.AccsSerialized))
			for i := range result.AccsSerialized {
				result.Accs[i] = account.NewAccount(publicKey, result.Tokens[i])
				if err = result.Accs[i].Deserialize(helpers.NewBufferReader(result.AccsSerialized[i])); err != nil {
					return nil, err
				}
			}
			result.AccsSerialized = nil

			if result.PlainAccSerialized != nil {
				result.PlainAcc = &plain_account.PlainAccount{}
				if err = result.PlainAcc.Deserialize(helpers.NewBufferReader(result.PlainAccSerialized)); err != nil {
					return nil, err
				}
				result.PlainAccSerialized = nil
			}

			if result.RegSerialized != nil {
				result.Reg = &registration.Registration{}
				if err = result.Reg.Deserialize(helpers.NewBufferReader(result.RegSerialized)); err != nil {
					return nil, err
				}
				result.RegSerialized = nil
			}

		}

		return convertJSONBytes(result)
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

		data := socket.SendJSONAwaitAnswer([]byte("account/txs"), &api_types.APIAccountTxsRequest{api_types.APIAccountBaseRequest{"", hash}, uint64(args[1].Int())}, 0)
		if data.Err != nil {
			return nil, data.Err
		}

		return convertBytes(data.Out)
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

		data := socket.SendJSONAwaitAnswer([]byte("account/mem-pool"), &api_types.APIAccountBaseRequest{"", hash}, 0)
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		result := make([]helpers.HexBytes, 0)
		if err := json.Unmarshal(data.Out, &result); err != nil {
			return nil, err
		}

		return convertJSONBytes(result)
	})
}

func getNetworkTransaction(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		txMutex.Lock()
		defer txMutex.Unlock()

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

		data := socket.SendJSONAwaitAnswer([]byte("tx"), &api_types.APIBlockCompleteRequest{height, hash, api_types.RETURN_SERIALIZED}, 0)
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

		return convertJSONBytes(received)
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
		data := socket.SendJSONAwaitAnswer([]byte("token-info"), &api_types.APITokenInfoRequest{hash}, 0)
		if data.Err != nil {
			return nil, data.Err
		}
		return data.Out, nil
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
		data := socket.SendJSONAwaitAnswer([]byte("token"), &api_types.APITokenRequest{hash, api_types.RETURN_SERIALIZED}, 0)
		if data.Err != nil {
			return nil, data.Err
		}
		tok := &token.Token{}
		if err = tok.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return nil, err
		}
		return convertJSONBytes(tok)
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

		data := socket.SendJSONAwaitAnswer([]byte("mem-pool"), &api_types.APIMempoolRequest{chainHash, args[1].Int(), args[2].Int()}, 0)
		if data.Err != nil {
			return nil, data.Err
		}

		return convertBytes(data.Out)
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
		data := socket.SendJSONAwaitAnswer([]byte("sub"), req, 0)
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

		data := socket.SendJSONAwaitAnswer([]byte("unsub"), &api_types.APIUnsubscriptionRequest{key, api_types.SubscriptionType(args[1].Int())}, 0)
		return true, data.Err
	})
}

func init() {
	txMutex = sync.Mutex{}
}
