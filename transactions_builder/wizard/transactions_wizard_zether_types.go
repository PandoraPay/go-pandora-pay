package wizard

import (
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type WizardZetherPayloadExtraDelegateStake struct {
	WizardZetherPayloadExtra `json:"-"`
	DelegatePublicKey        helpers.HexBytes                                        `json:"delegatePublicKey"`
	ConvertToUnclaimed       bool                                                    `json:"convertToUnclaimed"`
	DelegatedStakingUpdate   *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
	DelegatePrivateKey       helpers.HexBytes                                        `json:"delegatePrivateKey,omitempty"`
}

type WizardZetherPayloadExtraClaimStake struct {
	WizardZetherPayloadExtra `json:"-"`
	DelegatePrivateKey       helpers.HexBytes `json:"delegatePrivateKey"`
}

type WizardZetherPayloadExtraAssetCreate struct {
	WizardZetherPayloadExtra `json:"-"`
	Asset                    *asset.Asset `json:"asset"`
}

type WizardZetherPayloadExtraAssetSupplyIncrease struct {
	WizardZetherPayloadExtra `json:"-"`
	AssetId                  helpers.HexBytes `json:"assetId"`
	ReceiverPublicKey        helpers.HexBytes `json:"receiverPublicKey"`
	Value                    uint64           `json:"value"`
	AssetSupplyPrivateKey    helpers.HexBytes `json:"assetSupplyPublicKey"`
}

type WizardZetherPayloadExtra interface {
}

type WizardZetherTransfer struct {
	Asset              helpers.HexBytes
	From               []byte //private key
	FromBalanceDecoded uint64
	Destination        string
	Amount             uint64
	Burn               uint64
	Data               *TransactionsWizardData
	PayloadExtra       WizardZetherPayloadExtra
}

type WizardZetherPublicKeyIndex struct {
	Registered            bool
	RegisteredIndex       uint64
	RegistrationSignature []byte
}
