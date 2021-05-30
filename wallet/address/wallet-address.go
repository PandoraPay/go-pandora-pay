package wallet_address

import (
	"bytes"
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type WalletAddress struct {
	Version        Version                      `json:"version"`
	Name           string                       `json:"name"`
	SeedIndex      uint32                       `json:"seedIndex"`
	IsMine         bool                         `json:"isMine"`
	PrivateKey     *addresses.PrivateKey        `json:"privateKey"`
	PublicKey      helpers.HexBytes             `json:"publicKey"`
	PublicKeyHash  helpers.HexBytes             `json:"publicKeyHash"`
	AddressEncoded string                       `json:"addressEncoded"`
	DelegatedStake *WalletAddressDelegatedStake `json:"delegatedStake"`
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
