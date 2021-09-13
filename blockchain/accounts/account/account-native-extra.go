package account

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type AccountNativeExtra struct {
	helpers.SerializableInterface `json:"-"`
	account                       *Account
	Nonce                         uint64               `json:"nonce"`
	DelegatedStakeVersion         uint64               `json:"delegatedStakeVersion"`
	DelegatedStake                *dpos.DelegatedStake `json:"delegatedStake"`
}

func (accountNativeExtra *AccountNativeExtra) Validate() error {
	if accountNativeExtra.DelegatedStakeVersion > 1 {
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (accountNativeExtra *AccountNativeExtra) HasDelegatedStake() bool {
	return accountNativeExtra.DelegatedStakeVersion == 1
}

func (accountNativeExtra *AccountNativeExtra) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &accountNativeExtra.Nonce, 1)
}

func (accountNativeExtra *AccountNativeExtra) RefreshDelegatedStake(blockHeight uint64) (err error) {

	if accountNativeExtra.DelegatedStakeVersion == 0 {
		return
	}
	if !bytes.Equal(accountNativeExtra.account.Token, config.NATIVE_TOKEN_FULL) {
		return errors.New("Token is native token")
	}

	for i := len(accountNativeExtra.DelegatedStake.StakesPending) - 1; i >= 0; i-- {
		stakePending := accountNativeExtra.DelegatedStake.StakesPending[i]
		if stakePending.ActivationHeight <= blockHeight {

			if stakePending.PendingType == dpos.DelegatedStakePendingStake {
				if err = helpers.SafeUint64Add(&accountNativeExtra.DelegatedStake.StakeAvailable, stakePending.PendingAmount); err != nil {
					return
				}
			} else {
				if err = accountNativeExtra.account.AddBalanceUint(stakePending.PendingAmount); err != nil {
					return
				}
			}
			accountNativeExtra.DelegatedStake.StakesPending = append(accountNativeExtra.DelegatedStake.StakesPending[:i], accountNativeExtra.DelegatedStake.StakesPending[i+1:]...)
		}
	}

	if accountNativeExtra.DelegatedStake.IsDelegatedStakeEmpty() {
		accountNativeExtra.DelegatedStakeVersion = 0
		accountNativeExtra.DelegatedStake = nil
	}
	return
}

func (accountNativeExtra *AccountNativeExtra) GetDelegatedStakeAvailable() uint64 {
	if accountNativeExtra.DelegatedStakeVersion == 0 {
		return 0
	}
	return accountNativeExtra.DelegatedStake.GetDelegatedStakeAvailable()
}

func (accountNativeExtra *AccountNativeExtra) ComputeDelegatedStakeAvailable(chainHeight uint64) (uint64, error) {
	if accountNativeExtra.DelegatedStakeVersion == 0 {
		return 0, nil
	}
	return accountNativeExtra.DelegatedStake.ComputeDelegatedStakeAvailable(chainHeight)
}

func (accountNativeExtra *AccountNativeExtra) ComputeDelegatedUnstakePending() (uint64, error) {
	if accountNativeExtra.DelegatedStakeVersion == 0 {
		return 0, nil
	}
	return accountNativeExtra.DelegatedStake.ComputeDelegatedUnstakePending()
}

func (accountNativeExtra *AccountNativeExtra) CreateDelegatedStake(amount uint64, delegatedStakePublicKey []byte, delegatedStakeFee uint64) error {

	if accountNativeExtra.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}
	if delegatedStakePublicKey == nil || len(delegatedStakePublicKey) != cryptography.PublicKeySize {
		return errors.New("delegatedStakePublicKey is Invalid")
	}
	accountNativeExtra.DelegatedStakeVersion = 1
	accountNativeExtra.DelegatedStake = &dpos.DelegatedStake{
		StakeAvailable:     amount,
		StakesPending:      []*dpos.DelegatedStakePending{},
		DelegatedPublicKey: delegatedStakePublicKey,
		DelegatedStakeFee:  delegatedStakeFee,
	}

	return nil
}

func (accountNativeExtra *AccountNativeExtra) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(accountNativeExtra.Nonce)
	w.WriteUvarint(accountNativeExtra.DelegatedStakeVersion)

	if accountNativeExtra.DelegatedStakeVersion == 1 {
		accountNativeExtra.DelegatedStake.Serialize(w)
	}
}

func (accountNativeExtra *AccountNativeExtra) Deserialize(r *helpers.BufferReader) (err error) {

	if accountNativeExtra.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}
	if accountNativeExtra.DelegatedStakeVersion, err = r.ReadUvarint(); err != nil {
		return
	}

	if accountNativeExtra.DelegatedStakeVersion == 1 {
		accountNativeExtra.DelegatedStake = new(dpos.DelegatedStake)
		if err = accountNativeExtra.DelegatedStake.Deserialize(r); err != nil {
			return
		}
	}

	return
}
