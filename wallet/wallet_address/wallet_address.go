package wallet_address

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type WalletAddress struct {
	Version                    Version                                 `json:"version" msgpack:"version"`
	Name                       string                                  `json:"name" msgpack:"name"`
	SeedIndex                  uint32                                  `json:"seedIndex" msgpack:"seedIndex"`
	IsMine                     bool                                    `json:"isMine" msgpack:"isMine"`
	PrivateKey                 *addresses.PrivateKey                   `json:"privateKey" msgpack:"privateKey"`
	Registration               helpers.HexBytes                        `json:"registration" msgpack:"registration"`
	PublicKey                  helpers.HexBytes                        `json:"publicKey" msgpack:"publicKey"`
	BalancesDecoded            map[string]*WalletAddressBalanceDecoded `json:"balancesDecoded" msgpack:"balancesDecoded"`
	AddressEncoded             string                                  `json:"addressEncoded" msgpack:"addressEncoded"`
	AddressRegistrationEncoded string                                  `json:"addressRegistrationEncoded" msgpack:"addressRegistrationEncoded"`
	DelegatedStake             *WalletAddressDelegatedStake            `json:"delegatedStake" msgpack:"delegatedStake"`
}

func (addr *WalletAddress) GetDelegatedStakePrivateKey() []byte {
	if addr.DelegatedStake != nil {
		return addr.DelegatedStake.PrivateKey.Key
	}
	return nil
}

func (addr *WalletAddress) GetDelegatedStakePublicKey() []byte {
	if addr.DelegatedStake != nil {
		return addr.DelegatedStake.PublicKey
	}
	return nil
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

func (addr *WalletAddress) UpdateDecodedBalance(newPreviousValue uint64, assetId []byte) {
	found := addr.BalancesDecoded[hex.EncodeToString(assetId)]
	if found != nil {
		found.AmountDecoded = newPreviousValue
	} else {
		addr.BalancesDecoded[hex.EncodeToString(assetId)] = &WalletAddressBalanceDecoded{
			newPreviousValue, assetId,
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

func (addr *WalletAddress) Clone() *WalletAddress {

	if addr == nil {
		return nil
	}

	balancesDecoded := make(map[string]*WalletAddressBalanceDecoded)
	for k, v := range addr.BalancesDecoded {
		balancesDecoded[k] = v
	}

	return &WalletAddress{
		addr.Version,
		addr.Name,
		addr.SeedIndex,
		addr.IsMine,
		addr.PrivateKey,
		addr.Registration,
		addr.PublicKey,
		balancesDecoded,
		addr.AddressEncoded,
		addr.AddressRegistrationEncoded,
		addr.DelegatedStake,
	}
}
