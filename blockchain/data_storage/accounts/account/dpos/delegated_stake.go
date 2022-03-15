package dpos

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Version                       DelegatedStakeVersion `json:"version" msgpack:"version"`
	SpendPublicKey                []byte                `json:"spendPublicKey" msgpack:"spendPublicKey"`
}

func (dstake *DelegatedStake) Validate() error {
	switch dstake.Version {
	case NO_STAKING, STAKING:
		if len(dstake.SpendPublicKey) != 0 {
			return errors.New("Spend Public Key is invalid")
		}
	case STAKING_SPEND_REQUIRED:
		if len(dstake.SpendPublicKey) != cryptography.PublicKeySize {
			return errors.New("Spend Public Key length must be 33")
		}
	default:
		return errors.New("Invalid DelegatedStakeVersion version")
	}
	return nil
}

func (dstake *DelegatedStake) HasDelegatedStake() bool {
	return dstake != nil && (dstake.Version == STAKING || dstake.Version == STAKING_SPEND_REQUIRED)
}

func (dstake *DelegatedStake) CreateDelegatedStake(spendPublicKey []byte) error {

	if dstake.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}

	if len(spendPublicKey) > 0 {
		if len(spendPublicKey) != cryptography.PublicKeySize {
			return errors.New("SpendPublicKey size is invalid")
		}
		dstake.Version = STAKING_SPEND_REQUIRED
		dstake.SpendPublicKey = spendPublicKey
	} else {
		dstake.Version = STAKING
	}

	return nil
}

func (dstake *DelegatedStake) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(uint64(dstake.Version))
	if dstake.Version == STAKING_SPEND_REQUIRED {
		w.Write(dstake.SpendPublicKey)
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
	case STAKING_SPEND_REQUIRED:
		if dstake.SpendPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return errors.New("Spend Public Key is missing")
		}
	default:
		return errors.New("Invalid DelegatedStake version")
	}

	return
}
