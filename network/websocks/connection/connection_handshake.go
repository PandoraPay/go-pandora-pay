package connection

import (
	"errors"
	"github.com/blang/semver/v4"
	"pandora-pay/config"
	"pandora-pay/network/network_config"
)

type ConnectionHandshake struct {
	Name      string                       `json:"name" msgpack:"name"`
	Version   string                       `json:"version" msgpack:"version"`
	Network   uint64                       `json:"network" msgpack:"network"`
	Consensus network_config.ConsensusType `json:"consensus" msgpack:"consensus"`
	URL       string                       `json:"url" msgpack:"url"`
}

func (handshake *ConnectionHandshake) ValidateHandshake() (*semver.Version, error) {

	if handshake.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Network is different")
	}
	if handshake.Consensus >= network_config.CONSENSUS_TYPE_END {
		return nil, errors.New("Invalid CONSENSUS")
	}

	version, err := semver.Parse(handshake.Version)
	if err != nil {
		return nil, errors.New("Invalid VERSION format")
	}

	return &version, nil
}
