package connection

import (
	"errors"
	"net/url"
	"pandora-pay/config"
)

type ConnectionHandshake struct {
	Name      string               `json:"name"`
	Version   string               `json:"version"`
	Network   uint64               `json:"network"`
	Consensus config.ConsensusType `json:"consensus"`
	URLStr    string               `json:"url"`
	URL       *url.URL             `json:"-"`
}

func (handshake *ConnectionHandshake) ValidateHandshake() error {

	if handshake.Network != config.NETWORK_SELECTED {
		return errors.New("Network is different")
	}
	if handshake.Consensus >= config.CONSENSUS_TYPE_END {
		return errors.New("INVALID CONSENSUS")
	}

	var err error
	if handshake.URL, err = url.Parse(handshake.URLStr); err != nil {
		return err
	}

	return nil
}
