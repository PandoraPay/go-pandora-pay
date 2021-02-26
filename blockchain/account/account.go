package account

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain/account/dpos"
	"pandora-pay/helpers"
)

type Account struct {
	Version uint64
	Nonce   uint64

	Balances       []*Balance
	DelegatedStake *dpos.DelegatedStake
}

func (account *Account) HasDelegatedStake() bool {
	return account.Version == 1
}

func (account *Account) IsAccountEmpty() bool {
	return (account.Version == 0 && len(account.Balances) == 0) ||
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

func (account *Account) AddBalance(sign bool, amount uint64, currency []byte) error {

	var foundBalance *Balance
	var foundBalanceIndex int

	for i, balance := range account.Balances {
		if bytes.Equal(balance.Currency[:], currency[:]) {
			foundBalance = balance
			foundBalanceIndex = i
			break
		}
	}

	if sign {
		if foundBalance == nil {
			foundBalance = new(Balance)
			copy(foundBalance.Currency[:], currency[:])
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

func (account *Account) GetDelegatedStakeAvailable(blockHeight uint64) uint64 {
	if !account.HasDelegatedStake() {
		return 0
	}
	return account.DelegatedStake.GetDelegatedStakeAvailable(blockHeight)
}

func (account *Account) Serialize() []byte {

	var serialized bytes.Buffer
	temp := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(temp, account.Version)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, account.Nonce)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, uint64(len(account.Balances)))
	serialized.Write(temp[:n])

	for i := 0; i < len(account.Balances); i++ {
		account.Balances[i].Serialize(&serialized, temp)
	}

	if account.HasDelegatedStake() {
		account.DelegatedStake.Serialize(&serialized, temp)
	}

	return serialized.Bytes()
}

func (account *Account) Deserialize(buf []byte) (out []byte, err error) {

	if account.Version, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	if account.Nonce, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	var n uint64
	if n, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	for i := uint64(0); i < n; i++ {
		var balance = new(Balance)
		if buf, err = balance.Deserialize(buf); err != nil {
			return
		}
		account.Balances = append(account.Balances, balance)
	}

	if account.HasDelegatedStake() {
		account.DelegatedStake = new(dpos.DelegatedStake)
		if buf, err = account.DelegatedStake.Deserialize(buf); err != nil {
			return
		}
	}

	out = buf
	return
}
