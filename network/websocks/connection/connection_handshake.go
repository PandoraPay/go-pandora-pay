package connection

import (
	"errors"
	"pandora-pay/config"
)

type ConnectionHandshake struct {
	Name      string               `json:"name" msgpack:"name"`
	Version   string               `json:"version" msgpack:"version"`
	Network   uint64               `json:"network" msgpack:"network"`
	Consensus config.ConsensusType `json:"consensus" msgpack:"consensus"`
	URL       string               `json:"url" msgpack:"url"`
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
