package wallet_address

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/wallet/wallet_address/shared_staked"
)

type WalletAddress struct {
	Version                    Version                                  `json:"version" msgpack:"version"`
	Name                       string                                   `json:"name" msgpack:"name"`
	SeedIndex                  uint32                                   `json:"seedIndex" msgpack:"seedIndex"`
	IsMine                     bool                                     `json:"isMine" msgpack:"isMine"`
	SecretKey                  []byte                                   `json:"secretKey" msgpack:"secretKey"`
	PrivateKey                 *addresses.PrivateKey                    `json:"privateKey" msgpack:"privateKey"`
	SpendPrivateKey            *addresses.PrivateKey                    `json:"spendPrivateKey" msgpack:"spendPrivateKey"`
	Registration               []byte                                   `json:"registration" msgpack:"registration"`
	PublicKey                  []byte                                   `json:"publicKey" msgpack:"publicKey"`
	Staked                     bool                                     `json:"staked" msgpack:"staked"`
	SpendRequired              bool                                     `json:"spendRequired" msgpack:"spendRequired"`
	SpendPublicKey             []byte                                   `json:"spendPublicKey" msgpack:"spendPublicKey"`
	IsSharedStaked             bool                                     `json:"isSharedStaked,omitempty" msgpack:"isSharedStaked,omitempty"`
	SharedStaked               *shared_staked.WalletAddressSharedStaked `json:"sharedStaked,omitempty" msgpack:"sharedStaked,omitempty"`
	AddressEncoded             string                                   `json:"addressEncoded" msgpack:"addressEncoded"`
	AddressRegistrationEncoded string                                   `json:"addressRegistrationEncoded" msgpack:"addressRegistrationEncoded"`
}

func (addr *WalletAddress) DeriveSharedStaked() (*shared_staked.WalletAddressSharedStaked, error) {

	if addr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}

	return &shared_staked.WalletAddressSharedStaked{
		PrivateKey: addr.PrivateKey,
		PublicKey:  addr.PublicKey,
	}, nil

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

	var sharedStaked *shared_staked.WalletAddressSharedStaked
	if addr.SharedStaked != nil {
		sharedStaked = &shared_staked.WalletAddressSharedStaked{addr.SharedStaked.PrivateKey, addr.SharedStaked.PublicKey}
	}

	return &WalletAddress{
		addr.Version,
		addr.Name,
		addr.SeedIndex,
		addr.IsMine,
		addr.SecretKey,
		addr.PrivateKey,
		addr.SpendPrivateKey,
		addr.Registration,
		addr.PublicKey,
		addr.Staked,
		addr.SpendRequired,
		addr.SpendPublicKey,
		addr.IsSharedStaked,
		sharedStaked,
		addr.AddressEncoded,
		addr.AddressRegistrationEncoded,
	}
}
