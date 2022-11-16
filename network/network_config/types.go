package network_config

type ConsensusType uint8

const (
	CONSENSUS_TYPE_NONE ConsensusType = iota
	CONSENSUS_TYPE_FULL
	CONSENSUS_TYPE_WALLET

	CONSENSUS_TYPE_END
)
