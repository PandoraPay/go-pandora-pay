package api_websockets

import "pandora-pay/config"

type APIHandshake struct {
	Name      string
	Version   string
	Network   uint64
	Consensus config.ConsensusType
}

type APIBlockHeight = uint64
