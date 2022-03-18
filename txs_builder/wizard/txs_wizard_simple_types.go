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
