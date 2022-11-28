package config

type NodeConsensusType uint8

const (
	NODE_CONSENSUS_TYPE_NONE NodeConsensusType = iota
	NODE_CONSENSUS_TYPE_FULL
	NODE_CONSENSUS_TYPE_APP
)
