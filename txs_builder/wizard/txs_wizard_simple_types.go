package wizard

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
)

type WizardTxSimpleExtraUpdateDelegate struct {
	WizardTxSimpleExtra         `json:"-"  msgpack:"-"`
	DelegatedStakingClaimAmount uint64                                                  `json:"delegatedStakingClaimAmount"  msgpack:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"  msgpack:"delegatedStakingUpdate"`
}

type WizardTxSimpleExtraUnstake struct {
	WizardTxSimpleExtra `json:"-"  msgpack:"-"`
	Amount              uint64 `json:"amount"  msgpack:"amount"`
}

type WizardTxSimpleExtraUpdateAssetFeeLiquidity struct {
	WizardTxSimpleExtra `json:"-"  msgpack:"-"`
	Liquidities         []*asset_fee_liquidity.AssetFeeLiquidity `json:"liquidities"  msgpack:"liquidities"`
	CollectorHasNew     bool                                     `json:"collectorHasNew"  msgpack:"collectorHasNew"`
	Collector           []byte                                   `json:"collector"  msgpack:"collector"`
}

type WizardTxSimpleExtra interface {
}
