package account

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/blockchain/account/dpos"
	"pandora-pay/helpers"
)

type Account struct {
	Version   uint64
	Nonce     uint64
	PublicKey [20]byte

	Balances       []*Balance
	DelegatedStake *dpos.DelegatedStake
}

func (account *Account) Serialize() []byte {

	var serialized bytes.Buffer
	temp := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(temp, account.Version)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, account.Nonce)
	serialized.Write(temp[:n])

	serialized.Write(account.PublicKey[:])

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

	account.Version, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	account.Nonce, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	var data []byte
	data, buf, err = helpers.DeserializeBuffer(buf, 20)
	if err != nil {
		return
	}
	copy(account.PublicKey[:], data)

	var n uint64
	n, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	for i := uint64(0); i < n; i++ {
		var balance = new(Balance)
		buf, err = balance.Deserialize(buf)
		if err != nil {
			return
		}
		account.Balances = append(account.Balances, balance)
	}

	if account.HasDelegatedStake() {
		account.DelegatedStake = new(dpos.DelegatedStake)
		buf, err = account.DelegatedStake.Deserialize(buf)
		if err != nil {
			return
		}
	}

	out = buf
	return
}

func (account *Account) HasDelegatedStake() bool {
	return account.Version == 1
}

func (account *Account) IsAccountEmpty() bool {
	return (account.Version == 0 && len(account.Balances) == 0) ||
		(account.HasDelegatedStake() && account.DelegatedStake.IsDelegatedStakeEmpty())
}
