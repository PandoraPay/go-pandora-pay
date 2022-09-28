package main

import (
	"context"
	"encoding/base64"
	"pandora-pay/app"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"syscall/js"
	"time"
)

func networkDisconnect(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return app.Network.Websockets.Disconnect(), nil
	})
}

func getNetworkBlockchain(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_common.APIBlockchain](app.Network.Websockets.GetFirstSocket(), []byte("chain"), nil, nil, 0))
	})
}

func getNetworkFaucetCoins(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_faucet.APIFaucetCoinsReply](app.Network.Websockets.GetFirstSocket(), []byte("faucet/coins"), &api_faucet.APIFaucetCoinsRequest{args[0].String(), args[1].String()}, nil, 120*time.Second))
	})
}

func getNetworkFaucetInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_faucet.APIFaucetInfo](app.Network.Websockets.GetFirstSocket(), []byte("faucet/info"), nil, nil, 0))
	})
}

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIBlockInfoRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[info.BlockInfo](app.Network.Websockets.GetFirstSocket(), []byte("block-info"), request, nil, 0))
	})
}

func getNetworkBlockWithTxs(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIBlockRequest{0, nil, api_types.RETURN_SERIALIZED}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		blkWithTxs, err := connection.SendJSONAwaitAnswer[api_common.APIBlockReply](app.Network.Websockets.GetFirstSocket(), []byte("block"), request, nil, 0)
		if err != nil {
			return nil, err
		}

		blkWithTxs.Block = block.CreateEmptyBlock()
		if err := blkWithTxs.Block.Deserialize(helpers.NewBufferReader(blkWithTxs.BlockSerialized)); err != nil {
			return nil, err
		}
		if err := blkWithTxs.Block.BloomNow(); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(blkWithTxs)
	})
}

func getNetworkAccountsCount(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		assetId, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_common.APIAccountsCountReply](app.Network.Websockets.GetFirstSocket(), []byte("accounts/count"), &api_common.APIAccountsCountRequest{assetId}, nil, 0))
	})
}

func getNetworkAccount(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAccountRequest{api_types.APIAccountBaseRequest{}, api_types.RETURN_SERIALIZED}
		err := webassembly_utils.UnmarshalBytes(args[0], request)
		if err != nil {
			return nil, err
		}

		publicKey, err := request.GetPublicKeyHash(true)
		if err != nil {
			return nil, err
		}

		result, err := connection.SendJSONAwaitAnswer[api_common.APIAccountReply](app.Network.Websockets.GetFirstSocket(), []byte("account"), request, nil, 0)
		if err != nil {
			return nil, err
		}

		if result != nil {

			result.Accs = make([]*account.Account, len(result.AccsSerialized))
			for i := range result.AccsSerialized {
				if result.Accs[i], err = account.NewAccount(publicKey, result.AccsExtra[i].Index, result.AccsExtra[i].Asset); err != nil {
					return nil, err
				}
				if err = result.Accs[i].Deserialize(helpers.NewBufferReader(result.AccsSerialized[i])); err != nil {
					return nil, err
				}
			}
			result.AccsSerialized = nil

			if result.PlainAccSerialized != nil {
				result.PlainAcc = plain_account.NewPlainAccount(publicKey, result.PlainAccExtra.Index)
				if err = result.PlainAcc.Deserialize(helpers.NewBufferReader(result.PlainAccSerialized)); err != nil {
					return nil, err
				}
				result.PlainAccSerialized = nil
			}

		}

		return webassembly_utils.ConvertJSONBytes(result)
	})
}

func getNetworkAccountTxs(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAccountTxsRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_common.APIAccountTxsReply](app.Network.Websockets.GetFirstSocket(), []byte("account/txs"), request, nil, 0))
	})
}

func getNetworkAccountMempool(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_types.APIAccountBaseRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_common.APIAccountMempoolReply](app.Network.Websockets.GetFirstSocket(), []byte("account/mempool"), request, nil, 0))
	})
}

func getNetworkAccountMempoolNonce(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_types.APIAccountBaseRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_common.APIAccountMempoolNonceReply](app.Network.Websockets.GetFirstSocket(), []byte("account/mempool-nonce"), request, nil, 0))
	})
}

func getNetworkTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APITxRequest{0, nil, api_types.RETURN_SERIALIZED}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		received, err := connection.SendJSONAwaitAnswer[api_common.APITxReply](app.Network.Websockets.GetFirstSocket(), []byte("tx"), request, nil, 0)
		if err != nil {
			return nil, err
		}

		received.Tx = &transaction.Transaction{}
		if err := received.Tx.Deserialize(helpers.NewBufferReader(received.TxSerialized)); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(received)
	})
}

func getNetworkTxExists(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APITxExistsRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		received, err := connection.SendJSONAwaitAnswer[api_common.APITxExistsReply](app.Network.Websockets.GetFirstSocket(), []byte("tx/exists"), request, nil, 0)
		if err != nil {
			return nil, err
		}

		return received.Exists, nil
	})
}

func getNetworkBlockExists(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIBlockExistsRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		received, err := connection.SendJSONAwaitAnswer[api_common.APIBlockExistsReply](app.Network.Websockets.GetFirstSocket(), []byte("block/exists"), request, nil, 0)
		if err != nil {
			return nil, err
		}

		return received.Exists, nil
	})
}

func getNetworkTxPreview(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APITransactionPreviewRequest{0, nil}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		txPreviewReply, err := connection.SendJSONAwaitAnswer[api_common.APITransactionPreviewReply](app.Network.Websockets.GetFirstSocket(), []byte("tx-preview"), request, nil, 0)
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(txPreviewReply)
	})
}

func getNetworkAssetInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAssetInfoRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[info.AssetInfo](app.Network.Websockets.GetFirstSocket(), []byte("asset-info"), request, nil, 0))
	})
}

func getNetworkAsset(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAssetInfoRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		final, err := connection.SendJSONAwaitAnswer[api_common.APIAssetReply](app.Network.Websockets.GetFirstSocket(), []byte("asset"), &api_common.APIAssetRequest{request.Height, request.Hash, api_types.RETURN_SERIALIZED}, nil, 0)
		if err != nil {
			return nil, err
		}

		ast := asset.NewAsset(request.Hash, 0)
		if err = ast.Deserialize(helpers.NewBufferReader(final.Serialized)); err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(ast)
	})
}

func getNetworkMempool(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIMempoolRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertToJSONBytes(connection.SendJSONAwaitAnswer[api_common.APIMempoolReply](app.Network.Websockets.GetFirstSocket(), []byte("mempool"), request, nil, 0))
	})
}

func postNetworkMempoolBroadcastTransaction(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		tx := &transaction.Transaction{}
		if err := tx.Deserialize(helpers.NewBufferReader(webassembly_utils.GetBytes(args[0]))); err != nil {
			return nil, err
		}

		errs := app.Network.Websockets.BroadcastTxs([]*transaction.Transaction{tx}, true, true, advanced_connection_types.UUID_ALL, context.Background())
		if errs[0] != nil {
			return nil, errs[0]
		}

		return true, nil
	})
}

func subscribeNetwork(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		key, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		req := &api_types.APISubscriptionRequest{key, api_types.SubscriptionType(args[1].Int()), api_types.RETURN_SERIALIZED}
		_, err = connection.SendJSONAwaitAnswer[any](app.Network.Websockets.GetFirstSocket(), []byte("sub"), req, nil, 0)
		if err != nil {
			return nil, err
		}
		return true, nil
	})
}

func unsubscribeNetwork(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		key, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		_, err = connection.SendJSONAwaitAnswer[any](app.Network.Websockets.GetFirstSocket(), []byte("unsub"), &api_types.APIUnsubscriptionRequest{key, api_types.SubscriptionType(args[1].Int())}, nil, 0)
		if err != nil {
			return nil, err
		}
		return true, nil
	})
}
