package transaction_simple

type ScriptType uint64

const (
	SCRIPT_UPDATE_DELEGATE ScriptType = iota
	SCRIPT_UNSTAKE
	SCRIPT_END
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_UNSTAKE:
		return "SCRIPT_UNSTAKE"
	case SCRIPT_UPDATE_DELEGATE:
		return "SCRIPT_UPDATE_DELEGATE"
	default:
		return "Unknown ScriptType"
	}
}
