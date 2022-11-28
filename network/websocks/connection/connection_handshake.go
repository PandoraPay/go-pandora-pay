package connection

import (
	"errors"
	"github.com/blang/semver/v4"
	"pandora-pay/config"
)

type ConnectionHandshake struct {
	Name      string                   `json:"name" msgpack:"name"`
	Version   string                   `json:"version" msgpack:"version"`
	Network   uint64                   `json:"network" msgpack:"network"`
	Consensus config.NodeConsensusType `json:"consensus" msgpack:"consensus"`
	URL       string                   `json:"url" msgpack:"url"`
}

func (handshake *ConnectionHandshake) ValidateHandshake() (*semver.Version, error) {

	if handshake.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Network is different")
	}

	switch handshake.Consensus {
	case config.NODE_CONSENSUS_TYPE_NONE:
	case config.NODE_CONSENSUS_TYPE_FULL:
	case config.NODE_CONSENSUS_TYPE_APP:
	default:
		return nil, errors.New("Invalid CONSENSUS")
	}

	version, err := semver.Parse(handshake.Version)
	if err != nil {
		return nil, errors.New("Invalid VERSION format")
	}

	return &version, nil
}
