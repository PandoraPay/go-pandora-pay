package wallet_address

import (
	"errors"
	"pandora-pay/addresses"
)

type WalletAddress struct {
	Version        Version                    `json:"version" msgpack:"version"`
	Name           string                     `json:"name" msgpack:"name"`
	SeedIndex      uint32                     `json:"seedIndex" msgpack:"seedIndex"`
	IsMine         bool                       `json:"isMine" msgpack:"isMine"`
	SecretKey      []byte                     `json:"secretKey" msgpack:"secretKey"`
	PrivateKey     *addresses.PrivateKey      `json:"privateKey" msgpack:"privateKey"`
	PublicKey      []byte                     `json:"publicKey" msgpack:"publicKey"`
	PublicKeyHash  []byte                     `json:"publicKeyHash" msgpack:"publicKeyHash"`
	IsSharedStaked bool                       `json:"isSharedStaked,omitempty" msgpack:"isSharedStaked,omitempty"`
	SharedStaked   *WalletAddressSharedStaked `json:"sharedStaked,omitempty" msgpack:"sharedStaked,omitempty"`
	AddressEncoded string                     `json:"addressEncoded" msgpack:"addressEncoded"`
}

func (addr *WalletAddress) DeriveSharedStaked() (*WalletAddressSharedStaked, error) {

	if addr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}

	return &WalletAddressSharedStaked{
		PrivateKey: addr.PrivateKey,
		PublicKey:  addr.PublicKey,
	}, nil

}

func (addr *WalletAddress) GetAddress() string {
	return addr.AddressEncoded
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
	return addr.PrivateKey.Verify(message, signature), nil
}

func (addr *WalletAddress) Clone() *WalletAddress {

	if addr == nil {
		return nil
	}

	var sharedStaked *WalletAddressSharedStaked
	if addr.SharedStaked != nil {
		sharedStaked = &WalletAddressSharedStaked{addr.SharedStaked.PrivateKey, addr.SharedStaked.PublicKey}
	}

	return &WalletAddress{
		addr.Version,
		addr.Name,
		addr.SeedIndex,
		addr.IsMine,
		addr.SecretKey,
		addr.PrivateKey,
		addr.PublicKey,
		addr.PublicKeyHash,
		addr.IsSharedStaked,
		sharedStaked,
		addr.AddressEncoded,
	}
}
