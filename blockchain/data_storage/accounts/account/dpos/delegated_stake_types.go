package dpos

type DelegatedStakeVersion uint64

const (
	NO_STAKING DelegatedStakeVersion = iota
	STAKING
	STAKING_SPEND_REQUIRED
)

func (t DelegatedStakeVersion) String() string {
	switch t {
	case NO_STAKING:
		return "NO_STAKING"
	case STAKING:
		return "STAKING"
	case STAKING_SPEND_REQUIRED:
		return "STAKING_SPEND_REQUIRED"
	default:
		return "Unknown ScriptType"
	}
}
