package wizard

import (
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type ZetherTransferPayloadExtraDelegateStake struct {
	ZetherTransferPayloadExtra
	DelegatePublicKey      helpers.HexBytes                                        `json:"delegatePublicKey"`
	DelegatedStakingUpdate *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
	DelegatePrivateKey     helpers.HexBytes                                        `json:"delegatePrivateKey"`
}

type ZetherTransferPayloadExtraClaimStake struct {
	ZetherTransferPayloadExtra
	DelegatePrivateKey helpers.HexBytes `json:"delegatePrivateKey"`
}

type ZetherTransferPayloadExtraAssetCreate struct {
	ZetherTransferPayloadExtra
	Asset *asset.Asset `json:"asset"`
}

type ZetherTransferPayloadExtraAssetSupplyIncrease struct {
	ZetherTransferPayloadExtra
	AssetId              helpers.HexBytes `json:"assetId"`
	ReceiverPublicKey    helpers.HexBytes `json:"receiverPublicKey"`
	Value                uint64           `json:"value"`
	AssetSupplyPublicKey helpers.HexBytes `json:"assetSupplyPublicKey"`
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
