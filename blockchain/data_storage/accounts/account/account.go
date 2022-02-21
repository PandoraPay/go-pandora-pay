package account

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage/accounts/account/account_balance_homomorphic"
	"pandora-pay/blockchain/data_storage/accounts/account/dpos"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type Account struct {
	hash_map.HashMapElementSerializableInterface `json:"-" msgpack:"-"`
	PublicKey                                    []byte                                          `json:"-" msgpack:"-"` //hashmap key
	Asset                                        []byte                                          `json:"-" msgpack:"-"` //collection asset
	Index                                        uint64                                          `json:"-" msgpack:"-"` //hashmap Index
	Version                                      uint64                                          `json:"version" msgpack:"version"`
	Balance                                      *account_balance_homomorphic.BalanceHomomorphic `json:"balance" msgpack:"balance"`
	DelegatedStake                               *dpos.DelegatedStake                            `json:"delegatedStake" msgpack:"delegatedStake"`
}

func (account *Account) SetKey(key []byte) {
	account.PublicKey = key
}

func (account *Account) SetIndex(value uint64) {
	account.Index = value
}

func (account *Account) GetIndex() uint64 {
	return account.Index
}

func (account *Account) Validate() error {
	if account.Version != 0 {
		return errors.New("Version is invalid")
	}
	if bytes.Equal(account.Asset, config_coins.NATIVE_ASSET_FULL) {
		if account.DelegatedStake == nil {
			return errors.New("Delegated Stake must have not been nil ")
		}
		return account.DelegatedStake.Validate()
	} else {
		if account.DelegatedStake != nil {
			return errors.New("Delegated Stake must have been nil ")
		}
	}
	return nil
}

func (account *Account) GetBalance() (result *crypto.ElGamal) {
	return account.Balance.Amount
}

func (account *Account) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(account.Version)
	account.Balance.Serialize(w)
	if account.DelegatedStake != nil {
		account.DelegatedStake.Serialize(w)
	}
}

func (account *Account) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n != 0 {
		return errors.New("Invalid Account Version")
	}

	account.Version = n
	if err = account.Balance.Deserialize(r); err != nil {
		return
	}

	if account.DelegatedStake != nil {
		return account.Deserialize(r)
	}

	return
}

func (account *Account) RefreshDelegatedStake(blockHeight uint64) {
	if account.DelegatedStake != nil && account.DelegatedStake.HasDelegatedStake() {
		for i := len(account.DelegatedStake.StakesPending) - 1; i >= 0; i-- {
			stakePending := account.DelegatedStake.StakesPending[i]
			if stakePending.ActivationHeight <= blockHeight {
				account.Balance.AddEchanges(stakePending.PendingAmount.Amount)
				account.DelegatedStake.StakesPending = append(account.DelegatedStake.StakesPending[:i], account.DelegatedStake.StakesPending[i+1:]...)
			}
		}
	}
}

func NewAccount(publicKey []byte, index uint64, asset []byte) (*Account, error) {

	balance, err := account_balance_homomorphic.NewBalanceHomomorphicEmptyBalance(publicKey)
	if err != nil {
		return nil, err
	}

	acc := &Account{
		PublicKey: publicKey,
		Asset:     asset,
		Index:     index,
		Balance:   balance,
	}

	if bytes.Equal(asset, config_coins.NATIVE_ASSET_FULL) {
		acc.DelegatedStake = &dpos.DelegatedStake{Version: dpos.NO_STAKING}
	}

	return acc, nil
}

func NewAccountClear(publicKey []byte, index uint64, asset []byte) (*Account, error) {
	acc := &Account{
		PublicKey: publicKey,
		Asset:     asset,
		Index:     index,
		Balance:   &account_balance_homomorphic.BalanceHomomorphic{nil, nil},
	}

	if bytes.Equal(asset, config_coins.NATIVE_ASSET_FULL) {
		acc.DelegatedStake = &dpos.DelegatedStake{Version: dpos.NO_STAKING}
	}

	return acc, nil
}
