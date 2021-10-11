package plain_account

import (
	"errors"
	dpos "pandora-pay/blockchain/data_storage/plain_accounts/plain_account/dpos"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type PlainAccount struct {
	helpers.SerializableInterface `json:"-"`
	PublicKey                     []byte               `json:"-"`
	Nonce                         uint64               `json:"nonce"`
	Claimable                     uint64               `json:"claimable"`
	DelegatedStakeVersion         uint64               `json:"delegatedStakeVersion"`
	DelegatedStake                *dpos.DelegatedStake `json:"delegatedStake"`
}

func (plainAccount *PlainAccount) Validate() error {
	if plainAccount.DelegatedStakeVersion > 1 {
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (plainAccount *PlainAccount) HasDelegatedStake() bool {
	return plainAccount.DelegatedStakeVersion == 1
}

func (plainAccount *PlainAccount) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Nonce, 1)
}

func (plainAccount *PlainAccount) AddClaimable(sign bool, amount uint64) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Claimable, amount)
}

func (plainAccount *PlainAccount) RefreshDelegatedStake(blockHeight uint64) (err error) {

	if plainAccount.DelegatedStakeVersion == 0 {
		return
	}

	for i := len(plainAccount.DelegatedStake.StakesPending) - 1; i >= 0; i-- {
		stakePending := plainAccount.DelegatedStake.StakesPending[i]
		if stakePending.ActivationHeight <= blockHeight {

			if stakePending.PendingType == dpos.DelegatedStakePendingStake {
				if err = plainAccount.DelegatedStake.AddStakeAvailable(true, stakePending.PendingAmount); err != nil {
					return
				}
			} else {
				if err = plainAccount.AddClaimable(true, stakePending.PendingAmount); err != nil {
					return
				}
			}
			plainAccount.DelegatedStake.StakesPending = append(plainAccount.DelegatedStake.StakesPending[:i], plainAccount.DelegatedStake.StakesPending[i+1:]...)
		}
	}

	if plainAccount.DelegatedStake.IsDelegatedStakeEmpty() {
		plainAccount.DelegatedStakeVersion = 0
		plainAccount.DelegatedStake = nil
	}
	return
}

func (plainAccount *PlainAccount) GetDelegatedStakeAvailable() uint64 {
	if plainAccount.DelegatedStakeVersion == 0 {
		return 0
	}
	return plainAccount.DelegatedStake.GetDelegatedStakeAvailable()
}

func (plainAccount *PlainAccount) ComputeDelegatedStakeAvailable(chainHeight uint64) (uint64, error) {
	if plainAccount.DelegatedStakeVersion == 0 {
		return 0, nil
	}
	return plainAccount.DelegatedStake.ComputeDelegatedStakeAvailable(chainHeight)
}

func (plainAccount *PlainAccount) ComputeDelegatedUnstakePending() (uint64, error) {
	if plainAccount.DelegatedStakeVersion == 0 {
		return 0, nil
	}
	return plainAccount.DelegatedStake.ComputeDelegatedUnstakePending()
}

func (plainAccount *PlainAccount) CreateDelegatedStake(amount uint64, delegatedStakePublicKey []byte, delegatedStakeFee uint64) error {

	if plainAccount.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}
	if delegatedStakePublicKey == nil || len(delegatedStakePublicKey) != cryptography.PublicKeySize {
		return errors.New("delegatedStakePublicKey is Invalid")
	}
	plainAccount.DelegatedStakeVersion = 1
	plainAccount.DelegatedStake = &dpos.DelegatedStake{
		StakeAvailable:     amount,
		StakesPending:      []*dpos.DelegatedStakePending{},
		DelegatedPublicKey: delegatedStakePublicKey,
		DelegatedStakeFee:  delegatedStakeFee,
	}

	return nil
}

func (plainAccount *PlainAccount) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(plainAccount.Nonce)
	w.WriteUvarint(plainAccount.Claimable)
	w.WriteUvarint(plainAccount.DelegatedStakeVersion)

	if plainAccount.DelegatedStakeVersion == 1 {
		plainAccount.DelegatedStake.Serialize(w)
	}
}

func (plainAccount *PlainAccount) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	plainAccount.Serialize(w)
	return w.Bytes()
}

func (plainAccount *PlainAccount) Deserialize(r *helpers.BufferReader) (err error) {

	if plainAccount.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}
	if plainAccount.Claimable, err = r.ReadUvarint(); err != nil {
		return
	}
	if plainAccount.DelegatedStakeVersion, err = r.ReadUvarint(); err != nil {
		return
	}

	if plainAccount.DelegatedStakeVersion == 1 {
		plainAccount.DelegatedStake = new(dpos.DelegatedStake)
		if err = plainAccount.DelegatedStake.Deserialize(r); err != nil {
			return
		}
	}

	return
}

func NewPlainAccount(publicKey []byte) *PlainAccount {
	return &PlainAccount{
		PublicKey: publicKey,
	}
}
