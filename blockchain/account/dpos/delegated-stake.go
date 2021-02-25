package dpos

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type DelegatedStake struct {

	//public key for delegation
	DelegatedPublicKey [33]byte

	//confirmed stake
	StakeConfirmed uint64

	//when unstake can be done
	UnstakeHeight uint64

	//Pending stakes
	StakesPending []*DelegatedStakePending
}

func (delegatedStake *DelegatedStake) Serialize(serialized *bytes.Buffer, buf []byte) {

}

func (delegatedStake *DelegatedStake) Deserialize(buf []byte) (out []byte, err error) {

}
