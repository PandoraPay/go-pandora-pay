package dpos

import (
	"errors"
	"pandora-pay/helpers"
)

type DelegatedStake struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Version                       DelegatedStakeVersion `json:"version" msgpack:"version"`
	SpendPublicKey                []byte                `json:"spendPublicKey" msgpack:"spendPublicKey"`
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

func (dstake *DelegatedStake) CreateDelegatedStake(spendPublicKey []byte) error {
	if dstake.HasDelegatedStake() {
		return errors.New("It is already delegated")
	}

	dstake.Version = STAKING
	dstake.SpendPublicKey = spendPublicKey

	return nil
}

func (dstake *DelegatedStake) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(uint64(dstake.Version))
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
	default:
		return errors.New("Invalid DelegatedStake version")
	}

	return
}
