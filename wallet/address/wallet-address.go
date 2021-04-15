package wallet_address

import (
	"bytes"
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
)

type WalletAddress struct {
	Name           string
	SeedIndex      uint32
	IsMine         bool
	PrivateKey     *addresses.PrivateKey
	Address        *addresses.Address
	AddressEncoded string
	DelegatedStake *WalletAddressDelegatedStake
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

func (adr *WalletAddress) GetDelegatedStakePublicKeyHash() []byte {
	if adr.DelegatedStake != nil {
		return adr.DelegatedStake.PublicKeyHash
	}
	return nil
}

func (adr *WalletAddress) FindDelegatedStake(currentNonce, lastKnownNonce uint32, delegatedPublicKeyHash []byte) (delegatedStake *WalletAddressDelegatedStake, err error) {

	for nonce := lastKnownNonce; nonce <= currentNonce; nonce++ {

		if delegatedStake, err = adr.DeriveDelegatedStake(nonce); err != nil {
			return
		}
		if bytes.Equal(delegatedStake.PublicKeyHash, delegatedPublicKeyHash) {
			return
		}

	}

	return
}

func (adr *WalletAddress) DeriveDelegatedStake(nonce uint32) (delegatedStake *WalletAddressDelegatedStake, err error) {

	masterKey, err := bip32.NewMasterKey(adr.PrivateKey.Key)
	if err != nil {
		return
	}

	key, err := masterKey.NewChildKey(nonce)
	if err != nil {
		return
	}

	finalKey := cryptography.SHA3(key.Key)
	privateKey := &addresses.PrivateKey{Key: finalKey}

	address, err := privateKey.GenerateAddress(true, 0, []byte{})
	if err != nil {
		return
	}

	return &WalletAddressDelegatedStake{
		PrivateKey:     privateKey,
		PublicKeyHash:  address.PublicKeyHash,
		LastKnownNonce: nonce,
	}, nil
}
