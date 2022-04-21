package transaction_simple

type ScriptType uint64

const (
	SCRIPT_TRANSFER ScriptType = iota
	SCRIPT_UNSTAKE
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_TRANSFER:
		return "SCRIPT_TRANSFER"
	case SCRIPT_UNSTAKE:
		return "SCRIPT_UNSTAKE"
	default:
		return "Unknown ScriptType"
	}
}
