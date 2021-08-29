package account

import (
	"bytes"
	"errors"
	"math/big"
	"pandora-pay/blockchain/accounts/account/dpos"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type Account struct {
	helpers.SerializableInterface `json:"-"`
	PublicKey                     []byte                `json:"-"`
	Version                       uint64                `json:"version"`
	Nonce                         uint64                `json:"nonce"`
	BalancesHomo                  []*BalanceHomomorphic `json:"balancesHomo"`
	DelegatedStakeVersion         uint64                `json:"delegatedStakeVersion"`
	DelegatedStake                *dpos.DelegatedStake  `json:"delegatedStake"`
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

func (account *Account) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &account.Nonce, 1)
}

func (account *Account) AddBalanceHomoUint(amount uint64, tok []byte) (err error) {

	var foundBalance *BalanceHomomorphic

	for _, balance := range account.BalancesHomo {
		if bytes.Equal(balance.Token, tok) {
			foundBalance = balance
			break
		}
	}

	if foundBalance == nil {
		var acckey crypto.Point
		if err := acckey.DecodeCompressed(account.PublicKey); err != nil {
			panic(err)
		}
		foundBalance = &BalanceHomomorphic{crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G), tok}
		account.BalancesHomo = append(account.BalancesHomo, foundBalance)
	}

	foundBalance.Amount = foundBalance.Amount.Plus(new(big.Int).SetUint64(amount))

	return
}

func (account *Account) AddBalanceHomo(encryptedAmount []byte, tok []byte) (err error) {
	panic("not implemented")
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
				if err = account.AddBalanceHomoUint(stakePending.PendingAmount, config.NATIVE_TOKEN); err != nil {
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

func (account *Account) ComputeDelegatedUnstakePending() (uint64, error) {
	if account.DelegatedStakeVersion == 0 {
		return 0, nil
	}
	return account.DelegatedStake.ComputeDelegatedUnstakePending()
}

func (account *Account) GetBalanceHomo(token []byte) (result *crypto.ElGamal) {
	for _, balance := range account.BalancesHomo {
		if bytes.Equal(balance.Token, token) {
			return balance.Amount
		}
	}
	return nil
}

func (account *Account) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUvarint(account.Version)
	writer.WriteUvarint(account.Nonce)

	writer.WriteUvarint(uint64(len(account.BalancesHomo)))
	for _, balanceHomo := range account.BalancesHomo {
		balanceHomo.Serialize(writer)
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

func (account *Account) CreateDelegatedStake(amount uint64, delegatedStakePublicKey []byte, delegatedStakeFee uint64) error {
	if account.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}
	if delegatedStakePublicKey == nil || len(delegatedStakePublicKey) != cryptography.PublicKeySize {
		return errors.New("delegatedStakePublicKey is Invalid")
	}
	account.DelegatedStakeVersion = 1
	account.DelegatedStake = &dpos.DelegatedStake{
		StakeAvailable:     amount,
		StakesPending:      []*dpos.DelegatedStakePending{},
		DelegatedPublicKey: delegatedStakePublicKey,
		DelegatedStakeFee:  delegatedStakeFee,
	}

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
	account.BalancesHomo = make([]*BalanceHomomorphic, n)
	for i := uint64(0); i < n; i++ {
		var balance = new(BalanceHomomorphic)
		if err = balance.Deserialize(reader); err != nil {
			return
		}
		account.BalancesHomo[i] = balance
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
