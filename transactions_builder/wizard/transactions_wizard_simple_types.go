package wizard

import "pandora-pay/blockchain/transactions/transaction/transaction_data"

type TxTransferSimpleExtraUpdateDelegate struct {
	TxTransferSimpleExtra       `json:"-"`
	DelegatedStakingClaimAmount uint64                                                  `json:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
}

type TxTransferSimpleExtraUnstake struct {
	TxTransferSimpleExtra `json:"-"`
	Amount                uint64 `json:"amount"`
}

type TxTransferSimpleExtra interface {
}
