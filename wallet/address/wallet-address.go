package wallet_address

import (
	"bytes"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data/accounts/account"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type WalletAddress struct {
	Version                    Version                                 `json:"version"`
	Name                       string                                  `json:"name"`
	SeedIndex                  uint32                                  `json:"seedIndex"`
	IsMine                     bool                                    `json:"isMine"`
	PrivateKey                 *addresses.PrivateKey                   `json:"privateKey"`
	Registration               helpers.HexBytes                        `json:"registration"`
	PublicKey                  helpers.HexBytes                        `json:"publicKey"`
	BalancesDecoded            map[string]*WalletAddressBalanceDecoded `json:"balancesDecoded"`
	AddressEncoded             string                                  `json:"addressEncoded"`
	AddressRegistrationEncoded string                                  `json:"addressRegistrationEncoded"`
	DelegatedStake             *WalletAddressDelegatedStake            `json:"delegatedStake"`
}

func (adr *WalletAddress) GetDelegatedStakePrivateKey() []byte {
	if adr.DelegatedStake != nil {
		return adr.DelegatedStake.PrivateKey.Key
	}
	return nil
}

func (adr *WalletAddress) GetDelegatedStakePublicKey() []byte {
	if adr.DelegatedStake != nil {
		return adr.DelegatedStake.PublicKey
	}
	return nil
}

func (adr *WalletAddress) FindDelegatedStake(currentNonce, lastKnownNonce uint32, delegatedPublicKey []byte) (*WalletAddressDelegatedStake, error) {

	for nonce := lastKnownNonce; nonce <= currentNonce; nonce++ {

		delegatedStake, err := adr.DeriveDelegatedStake(nonce)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(delegatedStake.PublicKey, delegatedPublicKey) {
			return delegatedStake, nil
		}

	}

	return nil, errors.New("Nonce not found")
}

func (adr *WalletAddress) DeriveDelegatedStake(nonce uint32) (*WalletAddressDelegatedStake, error) {

	if adr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}

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

	address, err := privateKey.GenerateAddress(false, 0, []byte{})
	if err != nil {
		return nil, err
	}

	return &WalletAddressDelegatedStake{
		PrivateKey:     privateKey,
		PublicKey:      address.PublicKey,
		LastKnownNonce: nonce,
	}, nil
}

func (adr *WalletAddress) DecodeAccount(acc *account.Account, store bool) {

	if adr.PrivateKey == nil {
		return
	}

	if acc == nil {
		if store {
			adr.BalancesDecoded = make(map[string]*WalletAddressBalanceDecoded)
		}
		return
	}

	adr.DecodeBalance(acc.Balance.Amount, acc.Token, true)
}

func (adr *WalletAddress) DecodeBalance(balance *crypto.ElGamal, token []byte, store bool) uint64 {

	if adr.PrivateKey == nil {
		return 0
	}

	if len(token) == 0 {
		token = config.NATIVE_TOKEN_FULL
	}

	previousValue := uint64(0)
	found := adr.BalancesDecoded[string(token)]
	if found != nil {
		previousValue = found.AmountDecoded
	}

	newValue := adr.PrivateKey.DecodeBalance(balance, previousValue)

	if store {
		if found != nil {
			found.AmountDecoded = newValue
		} else {
			adr.BalancesDecoded[string(token)] = &WalletAddressBalanceDecoded{
				newValue, token,
			}
		}
	}

	return newValue
}

func (adr *WalletAddress) DecryptMessage(message []byte) ([]byte, error) {
	if adr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}
	return adr.PrivateKey.Decrypt(message)
}

func (adr *WalletAddress) SignMessage(message []byte) ([]byte, error) {
	if adr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}
	return adr.PrivateKey.Sign(message)
}
