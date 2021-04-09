package wallet_address

import (
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
)

type WalletAddress struct {
	Name           string
	SeedIndex      uint32
	Mine           bool
	PrivateKey     *addresses.PrivateKey
	Address        *addresses.Address
	DelegatedStake *WalletAddressDelegatedStaking
}

func (adr *WalletAddress) GetPublicKeyHash() []byte {
	return adr.Address.PublicKeyHash
}

func (adr *WalletAddress) GetAddressEncoded() string {
	return adr.Address.EncodeAddr()
}

func (adr *WalletAddress) GetDelegatedStakePrivateKey() []byte {

	if adr.DelegatedStake != nil {
		return adr.DelegatedStake.PrivateKey.Key
	}

	return nil
}

func (adr *WalletAddress) DeriveDelegatedStake(nonce uint32) (delegatedStaking *WalletAddressDelegatedStaking, err error) {

	masterKey, err := bip32.NewMasterKey(adr.PrivateKey.Key)
	if err != nil {
		return
	}

	key, err := masterKey.NewChildKey(nonce)
	if err != nil {
		return
	}

	finalKey := cryptography.SHA3(key.Key)

	delegatedStaking = &WalletAddressDelegatedStaking{
		PrivateKey: &addresses.PrivateKey{
			Key: finalKey,
		},
	}

	return
}
