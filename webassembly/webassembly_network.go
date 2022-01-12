package webassembly

import (
	"context"
	"encoding/hex"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/app"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
	"time"
)

func getNetworkBlockchain(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("chain"), nil, nil), &api_common.APIBlockchain{})
	})
}

func getNetworkFaucetCoins(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("faucet/coins"), &api_faucet.APIFaucetCoinsRequest{args[0].String(), args[1].String()}, ctx)
		if data.Err != nil {
			return nil, data.Err
		}
		return hex.EncodeToString(data.Out), nil
	})
}

func getNetworkFaucetInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("faucet/info"), nil, nil), &api_faucet.APIFaucetInfo{})
	})
}

func getNetworkBlockInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("block-info"), &api_common.APIBlockInfoRequest{uint64(args[0].Int()), hash}, nil), &info.BlockInfo{})
	})
}

func getNetworkBlockWithTxs(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("block"), &api_common.APIBlockRequest{uint64(args[0].Int()), hash, api_types.RETURN_SERIALIZED}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		blkWithTxs := &api_common.APIBlockWithTxsReply{}
		if err := msgpack.Unmarshal(data.Out, blkWithTxs); err != nil {
			return nil, err
		}

		blkWithTxs.Block = block.CreateEmptyBlock()
		if err := blkWithTxs.Block.Deserialize(helpers.NewBufferReader(blkWithTxs.BlockSerialized)); err != nil {
			return nil, err
		}
		if err = blkWithTxs.Block.BloomNow(); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(blkWithTxs)
	})
}

func getNetworkAccountsCount(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		assetId, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("accounts/count"), &api_common.APIAccountsCountRequest{assetId}, nil), &api_common.APIAccountsCountReply{})
	})
}

func getNetworkAccountsKeysByIndex(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAccountsKeysByIndexRequest{nil, nil, false}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("accounts/keys-by-index"), request, nil), &api_common.APIAccountsKeysByIndexReply{})
	})
}

func getNetworkAccountsByKeys(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAccountsByKeysRequest{nil, nil, false, api_types.RETURN_SERIALIZED}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("accounts/by-keys"), request, nil), &api_common.APIAccountsByKeysReply{})
	})
}

func getNetworkAccount(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_common.APIAccountRequest{api_types.APIAccountBaseRequest{}, api_types.RETURN_SERIALIZED}
		err := webassembly_utils.UnmarshalBytes(args[0], request)
		if err != nil {
			return nil, err
		}

		publicKey, err := request.GetPublicKey()
		if err != nil {
			return nil, err
		}

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("account"), request, nil)

		if data.Out == nil || data.Err != nil {
			return nil, data.Err
		}

		result := &api_common.APIAccount{}
		if err = msgpack.Unmarshal(data.Out, result); err != nil {
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

			if result.RegSerialized != nil {
				result.Reg = registration.NewRegistration(publicKey, result.RegExtra.Index)
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

		request := &api_common.APIAccountTxsRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("account/txs"), request, nil), &api_common.APIAccountTxsReply{})
	})
}

func getNetworkAccountMempool(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_types.APIAccountBaseRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("account/mempool"), request, nil), &api_common.APIAccountMempoolReply{})
	})
}

func getNetworkAccountMempoolNonce(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		request := &api_types.APIAccountBaseRequest{}
		if err := webassembly_utils.UnmarshalBytes(args[0], request); err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("account/mempool-nonce"), request, nil), &api_common.APIAccountMempoolNonceReply{})
	})
}

func getNetworkTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("tx"), &api_common.APIBlockCompleteRequest{uint64(args[0].Int()), hash, api_types.RETURN_SERIALIZED}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		received := &api_common.APITransactionReply{}
		if err = msgpack.Unmarshal(data.Out, received); err != nil {
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

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("tx-preview"), &api_common.APITransactionInfoRequest{uint64(args[0].Int()), hash}, nil)
		if data.Err != nil || data.Out == nil {
			return nil, data.Err
		}

		txPreviewReply := &api_common.APITransactionPreviewReply{}
		if err := msgpack.Unmarshal(data.Out, txPreviewReply); err != nil {
			return nil, err
		}

		switch txPreviewReply.TxPreview.Version {
		case transaction_type.TX_ZETHER:
			txPreviewReply.TxPreview.TxBase = &info.TxPreviewZether{}
		case transaction_type.TX_SIMPLE:
			txPreviewReply.TxPreview.TxBase = &info.TxPreviewSimple{}
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(data, txPreviewReply)
	})
}

func getNetworkAssetInfo(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("asset-info"), &api_common.APIAssetInfoRequest{uint64(args[0].Int()), hash}, nil), &api_common.APIAssetReply{})
	})
}

func getNetworkAsset(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("asset"), &api_common.APIAssetRequest{uint64(args[0].Int()), hash, api_types.RETURN_SERIALIZED}, nil)
		if data.Err != nil {
			return nil, data.Err
		}

		final := &api_common.APIAssetReply{}
		if err = msgpack.Unmarshal(data.Out, final); err != nil {
			return nil, err
		}

		ast := asset.NewAsset(nil, 0)
		if err = ast.Deserialize(helpers.NewBufferReader(final.Serialized)); err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(ast)
	})
}

func getNetworkMempool(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		chainHash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("mempool"), &api_common.APIMempoolRequest{chainHash, args[1].Int(), args[2].Int()}, nil), &api_common.APIMempoolReply{})
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

func getNetworkFeeLiquidity(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		hash, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertMsgPackToJSONBytes(app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("asset/fee-liquidity"), &api_common.APIAssetFeeLiquidityFeeRequest{uint64(args[0].Int()), hash}, nil), &api_common.APIAssetFeeLiquidityReply{})
	})
}

func subscribeNetwork(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		req := &api_types.APISubscriptionRequest{key, api_types.SubscriptionType(args[1].Int()), api_types.RETURN_SERIALIZED}
		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("sub"), req, nil)
		return true, data.Err
	})
}

func unsubscribeNetwork(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		data := app.Network.Websockets.GetFirstSocket().SendJSONAwaitAnswer([]byte("unsub"), &api_types.APIUnsubscriptionRequest{key, api_types.SubscriptionType(args[1].Int())}, nil)
		return true, data.Err
	})
}
