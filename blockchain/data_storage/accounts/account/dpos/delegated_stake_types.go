package dpos

type DelegatedStakeVersion uint64

const (
	NO_STAKING DelegatedStakeVersion = iota
	STAKING
)

func (t DelegatedStakeVersion) String() string {
	switch t {
	case NO_STAKING:
		return "NO_STAKING"
	case STAKING:
		return "STAKING"
	default:
		return "Unknown ScriptType"
	}
}
