package info

import transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"

type TxDetails struct {
	Version  transaction_type.TransactionVersion `json:"version"`
	TxScript uint64                              `json:"txScript"`
}
