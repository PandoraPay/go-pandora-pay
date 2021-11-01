package webassembly

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	api_faucet "pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/webassembly/webassembly_utils"
	"strconv"
	"syscall/js"
	"time"
)

func getNetworkBlockchain(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		data := socket.SendJSONAwaitAnswer([]byte("chain"), nil, nil)
		if data.Err != nil {
			return nil, data.Err
		}
		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkFaucetCoins(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		data := socket.SendJSONAwaitAnswer([]byte("faucet/coins"), &api_faucet.APIFaucetCoinsRequest{args[0].String(), args[1].String()}, ctx)
		if data.Err != nil {
			return nil, data.Err
		}
		return hex.EncodeToString(data.Out), nil
	})
}

func getNetworkFaucetInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		data := socket.SendJSONAwaitAnswer([]byte("faucet/info"), nil, nil)
		if data.Err != nil {
			return nil, data.Err
		}
		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}
		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}
		data := socket.SendJSONAwaitAnswer([]byte("block-info"), &api_types.APIBlockInfoRequest{api_types.APIHeightHash{uint64(args[0].Int()), hash}}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkBlockWithTxs(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("block"), &api_types.APIBlockRequest{api_types.APIHeightHash{uint64(args[0].Int()), hash}, api_types.RETURN_SERIALIZED}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		blkWithTxs := &api_types.APIBlockWithTxs{}
		if err := json.Unmarshal(data.Out, blkWithTxs); err != nil {
			return nil, err
		}

		blkWithTxs.Block = block.CreateEmptyBlock()
		if err := blkWithTxs.Block.Deserialize(helpers.NewBufferReader(blkWithTxs.BlockSerialized)); err != nil {
			return nil, err
		}
		if err = blkWithTxs.Block.BloomNow(); err != nil {
			return nil, err
		}

		blkWithTxs.BlockSerialized = nil

		return webassembly_utils.ConvertJSONBytes(blkWithTxs)
	})
}

func getNetworkAccountsCount(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		assetId, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendAwaitAnswer([]byte("accounts/count"), assetId, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return strconv.ParseUint(string(data.Out), 10, 64)
	})
}

func getNetworkAccountsKeysByIndex(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		request := &api_types.APIAccountsKeysByIndexRequest{nil, nil, false}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("accounts/keys-by-index"), request, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkAccountsByKeys(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		request := &api_types.APIAccountsByKeysRequest{nil, nil, false, api_types.RETURN_SERIALIZED}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("accounts/by-keys"), request, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkAccount(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account"), &api_types.APIAccountRequest{api_types.APIAccountBaseRequest{"", publicKey}, api_types.RETURN_SERIALIZED}, nil)
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
				if result.Accs[i], err = account.NewAccount(publicKey, result.Assets[i]); err != nil {
					return nil, err
				}
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

		return webassembly_utils.ConvertJSONBytes(result)
	})
}

func getNetworkAccountTxs(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account/txs"), &api_types.APIAccountTxsRequest{api_types.APIAccountBaseRequest{"", hash}, uint64(args[1].Int())}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkAccountMempool(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("account/mem-pool"), &api_types.APIAccountBaseRequest{"", hash}, nil)
		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		result := make([]helpers.HexBytes, 0)
		if err := json.Unmarshal(data.Out, &result); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(result)
	})
}

func getNetworkTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("tx"), &api_types.APIBlockCompleteRequest{api_types.APIHeightHash{uint64(args[0].Int()), hash}, api_types.RETURN_SERIALIZED}, nil)
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

		return webassembly_utils.ConvertJSONBytes(received)
	})
}

func getNetworkTxPreview(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("tx-preview"), &api_types.APITransactionInfoRequest{api_types.APIHeightHash{uint64(args[0].Int()), hash}}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkAssetInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("asset-info"), &api_types.APIAssetInfoRequest{api_types.APIHeightHash{uint64(args[0].Int()), hash}}, nil)
		if data.Err != nil {
			return nil, data.Err
		}
		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func getNetworkAsset(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("asset"), &api_types.APIAssetRequest{api_types.APIHeightHash{uint64(args[0].Int()), hash}, api_types.RETURN_SERIALIZED}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		ast := &asset.Asset{}
		if err = ast.Deserialize(helpers.NewBufferReader(data.Out)); err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(ast)
	})
}

func getNetworkMempool(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		chainHash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("mem-pool"), &api_types.APIMempoolRequest{chainHash, args[1].Int(), args[2].Int()}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		return webassembly_utils.ConvertBytes(data.Out), nil
	})
}

func postNetworkMempoolBroadcastTransaction(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		r := webassembly_utils.GetBytes(args[0])

		tx := &transaction.Transaction{}
		if err := tx.Deserialize(helpers.NewBufferReader(r)); err != nil {
			return nil, err
		}

		errs := app.Network.Consensus.BroadcastTxs([]*transaction.Transaction{tx}, true, false, advanced_connection_types.UUID_ALL, nil)
		if errs[0] != nil {
			return nil, errs[0]
		}

		return true, nil
	})
}

func subscribeNetwork(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		req := &api_types.APISubscriptionRequest{key, api_types.SubscriptionType(args[1].Int()), api_types.RETURN_SERIALIZED}
		data := socket.SendJSONAwaitAnswer([]byte("sub"), req, nil)
		return true, data.Err
	})
}

func unsubscribeNetwork(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := socket.SendJSONAwaitAnswer([]byte("unsub"), &api_types.APIUnsubscriptionRequest{key, api_types.SubscriptionType(args[1].Int())}, nil)
		return true, data.Err
	})
}

func init() {

}
