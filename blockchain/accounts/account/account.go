package account

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/helpers"
)

type Account struct {
	Version uint64
	Nonce   uint64

	Balances              []*Balance
	DelegatedStakeVersion uint64
	DelegatedStake        *dpos.DelegatedStake
}

func (account *Account) HasDelegatedStake() bool {
	return account.DelegatedStakeVersion == 1
}

func (account *Account) IsAccountEmpty() bool {
	return (!account.HasDelegatedStake() && len(account.Balances) == 0) ||
		(account.HasDelegatedStake() && account.DelegatedStake.IsDelegatedStakeEmpty())
}

func (account *Account) IncrementNonce(sign bool) error {

	if sign {
		account.Nonce += 1
	} else {
		if account.Nonce == 0 {
			return errors.New("Nonce would become negative")
		}
		account.Nonce -= 1
	}

	return nil
}

func (account *Account) AddBalance(sign bool, amount uint64, tok []byte) error {

	var foundBalance *Balance
	var foundBalanceIndex int

	for i, balance := range account.Balances {
		if bytes.Equal(balance.Token[:], tok[:]) {
			foundBalance = balance
			foundBalanceIndex = i
			break
		}
	}

	if sign {
		if foundBalance == nil {
			foundBalance = new(Balance)
			copy(foundBalance.Token[:], tok[:])
			account.Balances = append(account.Balances, foundBalance)
		}
		foundBalance.Amount += amount
	} else {

		if foundBalance == nil || foundBalance.Amount < amount {
			return errors.New("Balance doesn't exist or would become negative")
		}

		foundBalance.Amount -= amount
		if foundBalance.Amount == 0 {
			account.Balances = append(account.Balances[:foundBalanceIndex], account.Balances[:foundBalanceIndex+1]...)
		}

	}

	return nil
}

func (account *Account) AddReward(sign bool, amount, blockHeight uint64) {

	if !account.HasDelegatedStake() {
		panic("Strange. The accoun't doesn't have a delegated stake")
	}

	if sign {
		account.DelegatedStake.StakeAvailable += amount

	} else {
		if account.DelegatedStake.StakeAvailable < amount {
			panic("Strange. Stake available is less than reward. ")
		}
		account.DelegatedStake.StakeAvailable -= amount
	}

	account.refreshDelegatedStake(blockHeight)
}

func (account *Account) GetDelegatedStakeAvailable(blockHeight uint64) uint64 {
	if account.DelegatedStakeVersion == 0 {
		return 0
	}
	return account.DelegatedStake.GetDelegatedStakeAvailable(blockHeight)
}

func (account *Account) refreshDelegatedStake(blockHeight uint64) {
	account.DelegatedStake.RefreshDelegatedStake(blockHeight)
	if account.DelegatedStake.IsDelegatedStakeEmpty() {
		account.DelegatedStakeVersion = 0
		account.DelegatedStake = nil
	}
}

func (account *Account) Serialize() []byte {

	writer := helpers.NewBufferWriter()
	writer.WriteUint64(account.Version)
	writer.WriteUint64(account.Nonce)
	writer.WriteUint64(uint64(len(account.Balances)))

	for i := 0; i < len(account.Balances); i++ {
		account.Balances[i].Serialize(writer)
	}

	writer.WriteUint64(account.DelegatedStakeVersion)

	if account.DelegatedStakeVersion == 1 {
		account.DelegatedStake.Serialize(writer)
	}

	return writer.Bytes()
}

func (account *Account) Deserialize(buf []byte) (err error) {

	reader := helpers.NewBufferReader(buf)

	if account.Version, err = reader.ReadUvarint(); err != nil {
		return
	}
	if account.Version != 0 {
		err = errors.New("Version is invalid")
		return
	}

	if account.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}

	for i := uint64(0); i < n; i++ {
		var balance = new(Balance)
		if err = balance.Deserialize(reader); err != nil {
			return
		}
		account.Balances = append(account.Balances, balance)
	}

	if account.DelegatedStakeVersion, err = reader.ReadUvarint(); err != nil {
		return
	}
	if account.DelegatedStakeVersion > 1 {
		err = errors.New("Invalid DelegatedStakeVersion version")
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
