package wizard

import (
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type ZetherTransferPayloadExtraDelegateStake struct {
	DelegatePublicKey      []byte
	DelegatedStakingUpdate *transaction_data.TransactionDataDelegatedStakingUpdate
	DelegatePrivateKey     []byte
}

type ZetherTransferPayloadExtraClaimStake struct {
	DelegatePrivateKey helpers.HexBytes `json:"delegatePrivateKey"`
}

type ZetherTransferPayloadExtraAssetCreate struct {
	Asset *asset.Asset `json:"asset"`
}

type ZetherTransferPayloadExtra interface {
}

type ZetherTransfer struct {
	Asset              helpers.HexBytes
	From               []byte //private key
	FromBalanceDecoded uint64
	Destination        string
	Amount             uint64
	Burn               uint64
	Data               *TransactionsWizardData
	PayloadExtra       ZetherTransferPayloadExtra
}

type ZetherPublicKeyIndex struct {
	Registered            bool
	RegisteredIndex       uint64
	RegistrationSignature []byte
}
