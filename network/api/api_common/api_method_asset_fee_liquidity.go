package api_common

import (
	"encoding/binary"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAssetFeeLiquidityFeeRequest struct {
	Height uint64           `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   helpers.HexBytes `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

type APIAssetFeeLiquidityReply struct {
	Asset        helpers.HexBytes `json:"asset" msgpack:"asset"`
	Rate         uint64           `json:"rate" msgpack:"rate"`
	LeadingZeros byte             `json:"leadingZeros" msgpack:"leadingZeros"`
	Collector    helpers.HexBytes `json:"collector"  msgpack:"collector"` //collector Public Key
}

func (api *APICommon) AssetFeeLiquidity(r *http.Request, args *APIAssetFeeLiquidityFeeRequest, reply *APIAssetFeeLiquidityReply) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if args.Hash == nil {
			if args.Hash, err = api.ApiStore.loadAssetHash(reader, args.Height); err != nil {
				return
			}
		}

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		dataStorage := data_storage.NewDataStorage(reader)

		var plainAcc *plain_account.PlainAccount
		if plainAcc, err = dataStorage.GetWhoHasAssetTopLiquidity(args.Hash, chainHeight); err != nil || plainAcc == nil {
			return helpers.ReturnErrorIfNot(err, "Error retrieving Who Has Asset TopLiqiduity")
		}

		var liquditity *asset_fee_liquidity.AssetFeeLiquidity
		if liquditity = plainAcc.AssetFeeLiquidities.GetLiquidity(args.Hash); liquditity == nil {
			return errors.New("Error. It should have the liquidity")
		}

		reply.Asset = args.Hash
		reply.Rate = liquditity.Rate
		reply.LeadingZeros = liquditity.LeadingZeros
		reply.Collector = plainAcc.AssetFeeLiquidities.Collector

		return
	})
}

func (api *APICommon) GetAssetFeeLiquidity_http(values url.Values) (interface{}, error) {
	args := &APIAssetFeeLiquidityFeeRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIAssetFeeLiquidityReply{}
	return reply, api.AssetFeeLiquidity(nil, args, reply)
}

func (api *APICommon) GetAssetFeeLiquidity_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAssetFeeLiquidityFeeRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAssetFeeLiquidityReply{}
	return reply, api.AssetFeeLiquidity(nil, args, reply)
}
