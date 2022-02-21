package dpos

import (
	"errors"
	"pandora-pay/blockchain/data_storage/accounts/account/account_balance_homomorphic"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Version                       DelegatedStakeVersion    `json:"version" msgpack:"version"`
	SpendPublicKey                []byte                   `json:"spendPublicKey" msgpack:"spendPublicKey"`
	StakesPending                 []*DelegatedStakePending `json:"stakesPending,omitempty" msgpack:"stakesPending,omitempty"` //Pending stakes
}

func (dstake *DelegatedStake) Validate() error {
	switch dstake.Version {
	case NO_STAKING:
	case STAKING:
	default:
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (dstake *DelegatedStake) HasDelegatedStake() bool {
	return dstake.Version == STAKING
}

func (dstake *DelegatedStake) AddStakePendingStake(amount *crypto.ElGamal, blockHeight uint64) error {

	if !dstake.HasDelegatedStake() {
		return errors.New("plainAccount.HasDelegatedStake is false")
	}

	finalBlockHeight := blockHeight + config_stake.GetPendingStakeWindow(blockHeight)

	pending, err := account_balance_homomorphic.NewBalanceHomomorphic(amount)
	if err != nil {
		return err
	}

	dstake.StakesPending = append(dstake.StakesPending, &DelegatedStakePending{
		ActivationHeight: finalBlockHeight,
		PendingAmount:    pending,
	})

	return nil
}

func (dstake *DelegatedStake) ComputeDelegatedStakeAvailable(balance *crypto.ElGamal, blockHeight uint64) *crypto.ElGamal {
	if !dstake.HasDelegatedStake() {
		return nil
	}

	for i := range dstake.StakesPending {
		if dstake.StakesPending[i].ActivationHeight <= blockHeight {
			balance = balance.Add(dstake.StakesPending[i].PendingAmount.Amount)
		}
	}

	return balance
}

func (dstake *DelegatedStake) CreateDelegatedStake(spendPublicKey []byte) error {
	if dstake.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}

	dstake.Version = STAKING
	dstake.SpendPublicKey = spendPublicKey
	dstake.StakesPending = []*DelegatedStakePending{}

	return nil
}

func (dstake *DelegatedStake) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(uint64(dstake.Version))
	if dstake.Version == STAKING {

		w.WriteUvarint(uint64(len(dstake.StakesPending)))
		for _, stakePending := range dstake.StakesPending {
			stakePending.Serialize(w)
		}
	}
}

func (dstake *DelegatedStake) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	dstake.Version = DelegatedStakeVersion(n)

	switch dstake.Version {
	case NO_STAKING:
	case STAKING:

		if n, err = r.ReadUvarint(); err != nil {
			return
		}

		dstake.StakesPending = make([]*DelegatedStakePending, n)
		for i := uint64(0); i < n; i++ {
			delegatedStakePending := new(DelegatedStakePending)
			delegatedStakePending.PendingAmount = &account_balance_homomorphic.BalanceHomomorphic{nil, nil}
			if err = delegatedStakePending.Deserialize(r); err != nil {
				return
			}
			dstake.StakesPending[i] = delegatedStakePending
		}

	default:
		return errors.New("Invalid DelegatedStake version")
	}

	return
}
