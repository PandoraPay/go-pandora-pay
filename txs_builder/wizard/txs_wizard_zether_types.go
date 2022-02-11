package wizard

import (
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type WizardZetherPayloadExtraDelegateStake struct {
	WizardZetherPayloadExtra `json:"-" msgpack:"-"`
	DelegatePublicKey        helpers.HexBytes                                        `json:"delegatePublicKey" msgpack:"delegatePublicKey"`
	ConvertToUnclaimed       bool                                                    `json:"convertToUnclaimed" msgpack:"convertToUnclaimed"`
	DelegatedStakingUpdate   *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate" msgpack:"delegatedStakingUpdate"`
	DelegatePrivateKey       helpers.HexBytes                                        `json:"delegatePrivateKey,omitempty" msgpack:"delegatePrivateKey,omitempty"`
}

type WizardZetherPayloadExtraClaim struct {
	WizardZetherPayloadExtra `json:"-"  msgpack:""`
	DelegatePrivateKey       helpers.HexBytes `json:"delegatePrivateKey" msgpack:"delegatePrivateKey"`
}

type WizardZetherPayloadExtraAssetCreate struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
	Asset                    *asset.Asset `json:"asset" msgpack:"asset"`
}

type WizardZetherPayloadExtraAssetSupplyIncrease struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
	AssetId                  helpers.HexBytes `json:"assetId" msgpack:"assetId"`
	ReceiverPublicKey        helpers.HexBytes `json:"receiverPublicKey" msgpack:"receiverPublicKey"`
	Value                    uint64           `json:"value" msgpack:"value"`
	AssetSupplyPrivateKey    helpers.HexBytes `json:"assetSupplyPublicKey" msgpack:"assetSupplyPublicKey"`
}

type WizardZetherPayloadExtra interface {
}

type WizardZetherTransfer struct {
	Asset                  helpers.HexBytes         `json:"asset" msgpack:"asset"`
	Sender                 []byte                   `json:"sender" msgpack:"sender"` //private key
	SenderDecryptedBalance uint64                   `json:"senderDecryptedBalance" msgpack:"senderDecryptedBalance"`
	Recipient              string                   `json:"recipient" msgpack:"recipient"`
	Amount                 uint64                   `json:"amount" msgpack:"amount"`
	Burn                   uint64                   `json:"burn" msgpack:"burn"`
	FeeRate                uint64                   `json:"feeRate" msgpack:"feeRate"`
	FeeLeadingZeros        byte                     `json:"feeLeadingZeros" msgpack:"feeLeadingZeros"`
	Data                   *WizardTransactionData   `json:"data" msgpack:"data"`
	PayloadExtra           WizardZetherPayloadExtra `json:"payloadExtra" msgpack:"payloadExtra"`
}

type WizardZetherPublicKeyIndex struct {
	Registered            bool   `json:"registered" msgpack:"registered"`
	RegisteredIndex       uint64 `json:"registeredIndex" msgpack:"registeredIndex"`
	RegistrationSignature []byte `json:"registrationSignature" msgpack:"registrationSignature"`
}

type WizardZetherTransactionFee struct {
	*WizardTransactionFee
	Auto         bool   `json:"auto" msgpack:"auto"`
	Rate         uint64 `json:"rate" msgpack:"rate"`
	LeadingZeros byte   `json:"leadingZeros" msgpack:"leadingZeros"`
}
