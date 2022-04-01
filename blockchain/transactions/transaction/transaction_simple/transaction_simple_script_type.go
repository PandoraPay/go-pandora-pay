package transaction_simple

type ScriptType uint64

const (
	SCRIPT_TRANSFER ScriptType = iota
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_TRANSFER:
		return "SCRIPT_TRANSFER"
	default:
		return "Unknown ScriptType"
	}
}
