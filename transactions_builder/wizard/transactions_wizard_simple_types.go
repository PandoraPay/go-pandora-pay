package wizard

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type WizardTxSimpleExtraUpdateDelegate struct {
	WizardTxSimpleExtra         `json:"-"`
	DelegatedStakingClaimAmount uint64                                                  `json:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
}

type WizardTxSimpleExtraUnstake struct {
	WizardTxSimpleExtra `json:"-"`
	Amount              uint64 `json:"amount"`
}

type WizardTxSimpleExtraUpdateAssetFeeLiquidity struct {
	WizardTxSimpleExtra `json:"-"`
	Liquidities         []*asset_fee_liquidity.AssetFeeLiquidity
	CollectorHasNew     bool             `json:"CollectorHasNew"`
	Collector           helpers.HexBytes `json:"collector"`
}

type WizardTxSimpleExtra interface {
}
