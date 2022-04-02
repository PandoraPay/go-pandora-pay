package wizard

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
)

type WizardTxSimpleExtraUpdateAssetFeeLiquidity struct {
	WizardTxSimpleExtra `json:"-"  msgpack:"-"`
	Liquidities         []*asset_fee_liquidity.AssetFeeLiquidity `json:"liquidities"  msgpack:"liquidities"`
	CollectorHasNew     bool                                     `json:"collectorHasNew"  msgpack:"collectorHasNew"`
	Collector           []byte                                   `json:"collector"  msgpack:"collector"`
}

type WizardTxSimpleExtra interface {
}

type WizardTxSimpleTransfer struct {
	Extra  WizardTxSimpleExtra    `json:"extra" msgpack:"extra"`
	Data   *WizardTransactionData `json:"data" msgpack:"data"`
	Fee    *WizardTransactionFee  `json:"fee" msgpack:"fee"`
	Nonce  uint64                 `json:"nonce" msgpack:"nonce"`
	VinKey []byte                 `json:"vinKey" msgpack:"vinKey"`
}
