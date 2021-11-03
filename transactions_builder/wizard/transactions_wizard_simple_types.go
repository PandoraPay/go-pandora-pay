package wizard

import "pandora-pay/blockchain/transactions/transaction/transaction_data"

type WizardTxSimpleExtraUpdateDelegate struct {
	WizardTxSimpleExtra         `json:"-"`
	DelegatedStakingClaimAmount uint64                                                  `json:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
}

type WizardTxSimpleExtraUnstake struct {
	WizardTxSimpleExtra `json:"-"`
	Amount              uint64 `json:"amount"`
}

type WizardTxSimpleExtra interface {
}
