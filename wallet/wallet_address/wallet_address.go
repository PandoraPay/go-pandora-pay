package wallet_address

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
)

type WalletAddress struct {
	Version                    Version                                   `json:"version" msgpack:"version"`
	Name                       string                                    `json:"name" msgpack:"name"`
	SeedIndex                  uint32                                    `json:"seedIndex" msgpack:"seedIndex"`
	IsMine                     bool                                      `json:"isMine" msgpack:"isMine"`
	PrivateKey                 *addresses.PrivateKey                     `json:"privateKey" msgpack:"privateKey"`
	Registration               []byte                                    `json:"registration" msgpack:"registration"`
	PublicKey                  []byte                                    `json:"publicKey" msgpack:"publicKey"`
	DecryptedBalances          map[string]*WalletAddressDecryptedBalance `json:"decryptedBalances" msgpack:"decryptedBalances"`
	AddressEncoded             string                                    `json:"addressEncoded" msgpack:"addressEncoded"`
	AddressRegistrationEncoded string                                    `json:"addressRegistrationEncoded" msgpack:"addressRegistrationEncoded"`
}

func (addr *WalletAddress) FindDelegatedStake(currentNonce, lastKnownNonce uint32, delegatedStakePublicKey []byte) (*WalletAddressDelegatedStake, error) {

	for nonce := lastKnownNonce; nonce <= currentNonce; nonce++ {

		delegatedStake, err := addr.DeriveDelegatedStake(nonce)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(delegatedStake.PublicKey, delegatedStakePublicKey) {
			return delegatedStake, nil
		}

	}

	return nil, errors.New("Nonce not found")
}

func (addr *WalletAddress) DeriveDelegatedStake(nonce uint32) (*WalletAddressDelegatedStake, error) {

	if addr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}

	masterKey, err := bip32.NewMasterKey(addr.PrivateKey.Key)
	if err != nil {
		return nil, err
	}

	key, err := masterKey.NewChildKey(nonce)
	if err != nil {
		return nil, err
	}

	finalKey := cryptography.SHA3(key.Key)
	privateKey := &addresses.PrivateKey{Key: finalKey}

	address, err := privateKey.GenerateAddress(false, nil, 0, nil)
	if err != nil {
		return nil, err
	}

	return &WalletAddressDelegatedStake{
		PrivateKey:     privateKey,
		PublicKey:      address.PublicKey,
		LastKnownNonce: nonce,
	}, nil
}

func (addr *WalletAddress) UpdateDecryptedBalance(newDecryptedBalance uint64, balance []byte, assetId []byte) {
	found := addr.DecryptedBalances[base64.StdEncoding.EncodeToString(assetId)]
	if found != nil {
		found.Amount = newDecryptedBalance
	} else {
		addr.DecryptedBalances[base64.StdEncoding.EncodeToString(assetId)] = &WalletAddressDecryptedBalance{
			newDecryptedBalance,
			balance,
		}
	}
}

func (addr *WalletAddress) GetAddress(registered bool) string {
	if registered {
		return addr.AddressEncoded
	}
	return addr.AddressRegistrationEncoded
}

func (addr *WalletAddress) DecryptMessage(message []byte) ([]byte, error) {
	if addr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}
	return addr.PrivateKey.Decrypt(message)
}

func (addr *WalletAddress) SignMessage(message []byte) ([]byte, error) {
	if addr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}
	return addr.PrivateKey.Sign(message)
}

func (addr *WalletAddress) VerifySignedMessage(message, signature []byte) (bool, error) {
	address, err := addresses.DecodeAddr(addr.GetAddress(false))
	if err != nil {
		return false, err
	}
	return address.VerifySignedMessage(message, signature), nil
}

func (addr *WalletAddress) Clone() *WalletAddress {

	if addr == nil {
		return nil
	}

	decryptedBalances := make(map[string]*WalletAddressDecryptedBalance)
	for k, v := range addr.DecryptedBalances {
		decryptedBalances[k] = v
	}

	return &WalletAddress{
		addr.Version,
		addr.Name,
		addr.SeedIndex,
		addr.IsMine,
		addr.PrivateKey,
		addr.Registration,
		addr.PublicKey,
		decryptedBalances,
		addr.AddressEncoded,
		addr.AddressRegistrationEncoded,
	}
}
