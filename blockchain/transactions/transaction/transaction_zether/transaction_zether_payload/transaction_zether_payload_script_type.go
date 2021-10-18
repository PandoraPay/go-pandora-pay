package transaction_zether_payload

type PayloadScriptType uint64

const (
	SCRIPT_TRANSFER PayloadScriptType = iota
	SCRIPT_DELEGATE_STAKE
	SCRIPT_CLAIM_STAKE
	SCRIPT_END
)

func (t PayloadScriptType) String() string {
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
