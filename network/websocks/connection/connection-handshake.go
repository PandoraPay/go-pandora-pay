package connection

import (
	"errors"
	"pandora-pay/config"
)

type ConnectionHandshake struct {
	Name      string
	Version   string
	Network   uint64
	Consensus config.ConsensusType
}

func (handshake *ConnectionHandshake) ValidateHandshake() error {
	if handshake.Network != config.NETWORK_SELECTED {
		return errors.New("Network is different")
	}
	if handshake.Consensus >= config.CONSENSUS_TYPE_END {
		return errors.New("INVALID CONSENSUS")
	}
	return nil
}
