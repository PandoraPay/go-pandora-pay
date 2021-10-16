package transaction_zether

type ScriptType uint64

const (
	SCRIPT_TRANSFER ScriptType = iota
	SCRIPT_DELEGATE_STAKE
	SCRIPT_CLAIM_STAKE
	SCRIPT_END
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_TRANSFER:
		return "SCRIPT_TRANSFER"
	case SCRIPT_DELEGATE_STAKE:
		return "SCRIPT_DELEGATE_STAKE"
	case SCRIPT_CLAIM_STAKE:
		return "SCRIPT_CLAIM_STAKE"
	default:
		return "Unknown ScriptType"
	}
}
