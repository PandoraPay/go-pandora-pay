package wallet_address

import (
	"bytes"
	"errors"
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

func (adr *WalletAddress) FindDelegatedStake(currentNonce, lastKnownNonce uint32, delegatedPublicKeyHash []byte) (*WalletAddressDelegatedStake, error) {

	for nonce := lastKnownNonce; nonce <= currentNonce; nonce++ {

		delegatedStake, err := adr.DeriveDelegatedStake(nonce)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(delegatedStake.PublicKeyHash, delegatedPublicKeyHash) {
			return delegatedStake, nil
		}

	}

	return nil, errors.New("Nonce not found")
}

func (adr *WalletAddress) DeriveDelegatedStake(nonce uint32) (*WalletAddressDelegatedStake, error) {

	masterKey, err := bip32.NewMasterKey(adr.PrivateKey.Key)
	if err != nil {
		return nil, err
	}

	key, err := masterKey.NewChildKey(nonce)
	if err != nil {
		return nil, err
	}

	finalKey := cryptography.SHA3(key.Key)
	privateKey := &addresses.PrivateKey{Key: finalKey}

	address, err := privateKey.GenerateAddress(true, 0, []byte{})
	if err != nil {
		return nil, err
	}

	return &WalletAddressDelegatedStake{
		PrivateKey:     privateKey,
		PublicKeyHash:  address.PublicKeyHash,
		LastKnownNonce: nonce,
	}, nil
}
