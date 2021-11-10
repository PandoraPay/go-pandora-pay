package transaction_zether_payload

type PayloadScriptType uint64

const (
	SCRIPT_TRANSFER PayloadScriptType = iota
	SCRIPT_DELEGATE_STAKE
	SCRIPT_CLAIM
	SCRIPT_ASSET_CREATE
	SCRIPT_ASSET_SUPPLY_INCREASE
)

func (t PayloadScriptType) String() string {
	switch t {
	case SCRIPT_TRANSFER:
		return "SCRIPT_TRANSFER"
	case SCRIPT_DELEGATE_STAKE:
		return "SCRIPT_DELEGATE_STAKE"
	case SCRIPT_CLAIM:
		return "SCRIPT_CLAIM"
	case SCRIPT_ASSET_CREATE:
		return "SCRIPT_ASSET_CREATE"
	case SCRIPT_ASSET_SUPPLY_INCREASE:
		return "SCRIPT_ASSET_SUPPLY_INCREASE"
	default:
		return "Unknown ScriptType"
	}
}
