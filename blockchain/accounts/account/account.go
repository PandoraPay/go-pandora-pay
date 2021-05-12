package account

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Account struct {
	helpers.SerializableInterface
	Version               uint64
	Nonce                 uint64
	Balances              []*Balance
	DelegatedStakeVersion uint64
	DelegatedStake        *dpos.DelegatedStake
}

func (account *Account) Validate() error {
	if account.Version != 0 {
		return errors.New("Version is invalid")
	}
	if account.DelegatedStakeVersion > 1 {
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (account *Account) HasDelegatedStake() bool {
	return account.DelegatedStakeVersion == 1
}

func (account *Account) IsAccountEmpty() bool {
	return len(account.Balances) == 0 && (!account.HasDelegatedStake() || (account.HasDelegatedStake() && account.DelegatedStake.IsDelegatedStakeEmpty()))
}

func (account *Account) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &account.Nonce, 1)
}

func (account *Account) AddBalance(sign bool, amount uint64, tok []byte) (err error) {

	if amount == 0 {
		return
	}

	var foundBalance *Balance
	var foundBalanceIndex int

	for i, balance := range account.Balances {
		if bytes.Equal(balance.Token, tok) {
			foundBalance = balance
			foundBalanceIndex = i
			break
		}
	}

	if sign {
		if foundBalance == nil {
			foundBalance = &Balance{
				Token: tok,
			}
			account.Balances = append(account.Balances, foundBalance)
		}
		if err = helpers.SafeUint64Add(&foundBalance.Amount, amount); err != nil {
			return
		}
	} else {

		if foundBalance == nil {
			return errors.New("Balance doesn't exist or would become negative")
		}
		if err = helpers.SafeUint64Sub(&foundBalance.Amount, amount); err != nil {
			return
		}

		if foundBalance.Amount == 0 {
			//fast removal
			account.Balances[foundBalanceIndex] = account.Balances[len(account.Balances)-1]
			account.Balances = account.Balances[:len(account.Balances)-1]
		}

	}

	return
}

func (account *Account) RefreshDelegatedStake(blockHeight uint64) (err error) {
	if account.DelegatedStakeVersion == 0 {
		return
	}

	for i := len(account.DelegatedStake.StakesPending) - 1; i >= 0; i-- {
		stakePending := account.DelegatedStake.StakesPending[i]
		if stakePending.ActivationHeight <= blockHeight {

			if stakePending.PendingType == dpos.DelegatedStakePendingStake {
				if err = helpers.SafeUint64Add(&account.DelegatedStake.StakeAvailable, stakePending.PendingAmount); err != nil {
					return
				}
			} else {
				if err = account.AddBalance(true, stakePending.PendingAmount, config.NATIVE_TOKEN); err != nil {
					return
				}
			}
			account.DelegatedStake.StakesPending = append(account.DelegatedStake.StakesPending[:i], account.DelegatedStake.StakesPending[i+1:]...)
		}
	}

	if account.DelegatedStake.IsDelegatedStakeEmpty() {
		account.DelegatedStakeVersion = 0
		account.DelegatedStake = nil
	}
	return
}

func (account *Account) GetDelegatedStakeAvailable() uint64 {
	if account.DelegatedStakeVersion == 0 {
		return 0
	}
	return account.DelegatedStake.GetDelegatedStakeAvailable()
}

func (account *Account) ComputeDelegatedStakeAvailable(chainHeight uint64) (uint64, error) {
	if account.DelegatedStakeVersion == 0 {
		return 0, nil
	}
	return account.DelegatedStake.ComputeDelegatedStakeAvailable(chainHeight)
}

func (account *Account) GetAvailableBalance(token []byte) (result uint64, err error) {
	for _, balance := range account.Balances {
		if bytes.Equal(balance.Token, token) {
			result = balance.Amount
			break
		}
	}
	return
}

func (account *Account) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUvarint(account.Version)
	writer.WriteUvarint(account.Nonce)
	writer.WriteUvarint(uint64(len(account.Balances)))

	for i := 0; i < len(account.Balances); i++ {
		account.Balances[i].Serialize(writer)
	}

	writer.WriteUvarint(account.DelegatedStakeVersion)

	if account.DelegatedStakeVersion == 1 {
		account.DelegatedStake.Serialize(writer)
	}

}

func (account *Account) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	account.Serialize(writer)
	return writer.Bytes()
}

func (account *Account) CreateDelegatedStake(amount uint64, delegatedStakePublicKeyHash []byte) error {
	if account.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}
	if delegatedStakePublicKeyHash == nil || len(delegatedStakePublicKeyHash) != cryptography.KeyHashSize {
		return errors.New("DelegatedStakePublicKeyHash is Invalid")
	}
	account.DelegatedStakeVersion = 1
	account.DelegatedStake = new(dpos.DelegatedStake)
	account.DelegatedStake.StakeAvailable = amount
	account.DelegatedStake.DelegatedPublicKeyHash = delegatedStakePublicKeyHash

	return nil
}

func (account *Account) Deserialize(reader *helpers.BufferReader) (err error) {

	if account.Version, err = reader.ReadUvarint(); err != nil {
		return
	}
	if account.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	account.Balances = make([]*Balance, n)
	for i := uint64(0); i < n; i++ {
		var balance = new(Balance)
		if err = balance.Deserialize(reader); err != nil {
			return
		}
		account.Balances[i] = balance
	}

	if account.DelegatedStakeVersion, err = reader.ReadUvarint(); err != nil {
		return
	}
	if account.DelegatedStakeVersion == 1 {
		account.DelegatedStake = new(dpos.DelegatedStake)
		if err = account.DelegatedStake.Deserialize(reader); err != nil {
			return
		}
	}

	return
}
