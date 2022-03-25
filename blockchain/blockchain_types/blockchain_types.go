package blockchain_types

import (
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/transactions/transaction"
)

type BlockchainTransactionKeyUpdate struct {
	PublicKey []byte
	TxsCount  uint64
}

type BlockchainTransactionUpdate struct {
	TxHash         []byte
	TxHashStr      string
	Tx             *transaction.Transaction
	Inserted       bool
	BlockHeight    uint64
	BlockTimestamp uint64
	Height         uint64
	Keys           []*BlockchainTransactionKeyUpdate
}

type MempoolTransactionUpdate struct {
	Inserted                         bool
	Tx                               *transaction.Transaction
	IncludedInBlockchainNotification bool
	Keys                             map[string]bool
}

type BlockchainUpdates struct {
	AccsCollection *accounts.AccountsCollection
	PlainAccounts  *plain_accounts.PlainAccounts
	Assets         *assets.Assets
	Registrations  *registrations.Registrations
	BlockHeight    uint64
	BlockHash      []byte
}

type BlockchainSolutionAnswer struct {
	Err             error
	ChainKernelHash []byte
}

type BlockchainSolution struct {
	BlkComplete *block_complete.BlockComplete
	Done        chan *BlockchainSolutionAnswer
}
