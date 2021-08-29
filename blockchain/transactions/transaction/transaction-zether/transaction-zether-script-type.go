package transaction_zether

type ScriptType uint64

const (
	SCRIPT_TRANSFER ScriptType = iota
	SCRIPT_DELEGATE
	SCRIPT_END
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_TRANSFER:
		return "SCRIPT_TRANSFER"
	case SCRIPT_DELEGATE:
		return "SCRIPT_DELEGATE"
	default:
		return "Unknown ScriptType"
	}
}
