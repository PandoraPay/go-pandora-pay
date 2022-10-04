package wizard

import (
	"pandora-pay/blockchain/data_storage/assets/asset"
)

type WizardZetherPayloadExtraStaking struct {
	WizardZetherPayloadExtra `json:"-" msgpack:"-"`
}

type WizardZetherPayloadExtraStakingReward struct {
	WizardZetherPayloadExtra `json:"-"  msgpack:""`
	Reward                   uint64 `json:"reward" msgpack:"reward"`
}

type WizardZetherPayloadExtraAssetCreate struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
	Asset                    *asset.Asset `json:"asset" msgpack:"asset"`
}

type WizardZetherPayloadExtraSpend struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
}

type WizardZetherPayloadExtraAssetSupplyIncrease struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
	AssetId                  []byte `json:"assetId" msgpack:"assetId"`
	ReceiverPublicKey        []byte `json:"receiverPublicKey" msgpack:"receiverPublicKey"`
	Value                    uint64 `json:"value" msgpack:"value"`
	AssetSupplyPrivateKey    []byte `json:"assetSupplyPublicKey" msgpack:"assetSupplyPublicKey"`
}

type WizardZetherPayloadExtraPlainAccountFund struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
	PlainAccountPublicKey    []byte `json:"plainAccountPublicKey" msgpack:"plainAccountPublicKey"`
}

type WizardZetherPayloadExtraPayInFuture struct {
	WizardZetherPayloadExtra `json:"-" msgpack:""`
	Deadline                 uint64   `json:"deadline" msgpack:"deadline"`
	DefaultResolution        bool     `json:"defaultResolution" msgpack:"defaultResolution"`
	Threshold                byte     `json:"threshold" msgpack:"threshold"`
	MultisigPublicKeys       [][]byte `json:"multisigPublicKeys" msgpack:"multisigPublicKeys"`
}

type WizardZetherPayloadExtra interface {
}

type WizardZetherTransfer struct {
	Asset                  []byte                   `json:"asset" msgpack:"asset"`
	SenderPrivateKey       []byte                   `json:"senderPrivateKey" msgpack:"senderPrivateKey"` //private key
	SenderDecryptedBalance uint64                   `json:"senderDecryptedBalance" msgpack:"senderDecryptedBalance"`
	SenderSpendRequired    bool                     `json:"senderSpendRequired" msgpack:"senderSpendRequired"`
	SenderSpendPrivateKey  []byte                   `json:"senderSpendPrivateKey" msgpack:"senderSpendPrivateKey"`
	Recipient              string                   `json:"recipient" msgpack:"recipient"`
	Amount                 uint64                   `json:"amount" msgpack:"amount"`
	Burn                   uint64                   `json:"burn" msgpack:"burn"`
	FeeRate                uint64                   `json:"feeRate" msgpack:"feeRate"`
	FeeLeadingZeros        byte                     `json:"feeLeadingZeros" msgpack:"feeLeadingZeros"`
	Data                   *WizardTransactionData   `json:"data" msgpack:"data"`
	PayloadExtra           WizardZetherPayloadExtra `json:"payloadExtra" msgpack:"payloadExtra"`
	WitnessIndexes         []int                    `json:"witnessIndexes" msgpack:"witnessIndexes"`
}

type WizardZetherPublicKeyIndex struct {
	Registered                 bool   `json:"registered" msgpack:"registered"`
	RegisteredIndex            uint64 `json:"registeredIndex" msgpack:"registeredIndex"`
	RegistrationStaked         bool   `json:"registrationStaked" msgpack:"registrationStaked"`
	RegistrationSpendPublicKey []byte `json:"registrationSpendPublicKey" msgpack:"registrationSpendPublicKey"`
	RegistrationSignature      []byte `json:"registrationSignature" msgpack:"registrationSignature"`
}

type WizardZetherTransactionFee struct {
	*WizardTransactionFee
	Auto         bool   `json:"auto" msgpack:"auto"`
	Rate         uint64 `json:"rate" msgpack:"rate"`
	LeadingZeros byte   `json:"leadingZeros" msgpack:"leadingZeros"`
}
