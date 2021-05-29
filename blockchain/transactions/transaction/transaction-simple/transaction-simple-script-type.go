package transaction_simple

type ScriptType uint64

const (
	SCRIPT_NORMAL ScriptType = iota
	SCRIPT_UNSTAKE
	SCRIPT_WITHDRAW
	SCRIPT_DELEGATE

	SCRIPT_END
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_NORMAL:
		return "SCRIPT_NORMAL"
	case SCRIPT_UNSTAKE:
		return "SCRIPT_UNSTAKE"
	case SCRIPT_WITHDRAW:
		return "SCRIPT_WITHDRAW"
	case SCRIPT_DELEGATE:
		return "SCRIPT_DELEGATE"
	default:
		return "Unknown ScriptType"
	}
}
