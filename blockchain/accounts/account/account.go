package account

import (
	"bytes"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/config"
	"pandora-pay/helpers"
)

type Account struct {
	Version               uint64
	Nonce                 uint64
	Balances              []*Balance
	DelegatedStakeVersion uint64
	DelegatedStake        *dpos.DelegatedStake
}

func (account *Account) Validate() {
	if account.Version != 0 {
		panic("Version is invalid")
	}
	if account.DelegatedStakeVersion > 1 {
		panic("Invalid DelegatedStakeVersion version")
	}
}

func (account *Account) HasDelegatedStake() bool {
	return account.DelegatedStakeVersion == 1
}

func (account *Account) IsAccountEmpty() bool {
	return (!account.HasDelegatedStake() && len(account.Balances) == 0) ||
		(account.HasDelegatedStake() && account.DelegatedStake.IsDelegatedStakeEmpty())
}

func (account *Account) IncrementNonce(sign bool) {
	helpers.SafeUint64Update(sign, &account.Nonce, 1)
}

func (account *Account) AddBalance(sign bool, amount uint64, tok []byte) {

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
		helpers.SafeUint64Add(&foundBalance.Amount, amount)
	} else {

		if foundBalance == nil {
			panic("Balance doesn't exist or would become negative")
		}
		helpers.SafeUint64Sub(&foundBalance.Amount, amount)

		if foundBalance.Amount == 0 {
			//fast removal
			account.Balances[foundBalanceIndex] = account.Balances[len(account.Balances)-1]
			account.Balances = account.Balances[:len(account.Balances)-1]
		}

	}

}

func (account *Account) RefreshDelegatedStake(blockHeight uint64) {
	if account.DelegatedStakeVersion == 0 {
		return
	}

	for i := len(account.DelegatedStake.StakesPending) - 1; i >= 0; i-- {
		stakePending := account.DelegatedStake.StakesPending[i]
		if stakePending.ActivationHeight <= blockHeight {

			if stakePending.PendingType == true {
				helpers.SafeUint64Add(&account.DelegatedStake.StakeAvailable, stakePending.PendingAmount)
			} else {
				account.AddBalance(true, stakePending.PendingAmount, config.NATIVE_TOKEN)
			}
			account.DelegatedStake.StakesPending = append(account.DelegatedStake.StakesPending[:i], account.DelegatedStake.StakesPending[i+1:]...)
		}
	}

	if account.DelegatedStake.IsDelegatedStakeEmpty() {
		account.DelegatedStakeVersion = 0
		account.DelegatedStake = nil
	}

}

func (account *Account) GetDelegatedStakeAvailable(blockHeight uint64) uint64 {
	if account.DelegatedStakeVersion == 0 {
		return 0
	}
	return account.DelegatedStake.GetDelegatedStakeAvailable(blockHeight)
}

func (account *Account) GetAvailableBalance(blockHeight uint64, token []byte) (result uint64) {
	for _, balance := range account.Balances {
		if bytes.Equal(balance.Token, token) {
			result = balance.Amount
			if bytes.Equal(token, config.NATIVE_TOKEN) && account.DelegatedStakeVersion == 1 {
				helpers.SafeUint64Add(&result, account.DelegatedStake.GetDelegatedUnstakeAvailable(blockHeight))
			}
			return
		}
	}
	return
}

func (account *Account) Serialize() []byte {

	writer := helpers.NewBufferWriter()
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

	return writer.Bytes()
}

func (account *Account) Deserialize(buf []byte) {

	reader := helpers.NewBufferReader(buf)

	account.Version = reader.ReadUvarint()
	account.Nonce = reader.ReadUvarint()

	n := reader.ReadUvarint()
	account.Balances = make([]*Balance, n)
	for i := uint64(0); i < n; i++ {
		var balance = new(Balance)
		balance.Deserialize(reader)
		account.Balances[i] = balance
	}

	account.DelegatedStakeVersion = reader.ReadUvarint()
	if account.DelegatedStakeVersion == 1 {
		account.DelegatedStake = new(dpos.DelegatedStake)
		account.DelegatedStake.Deserialize(reader)
	}

}
