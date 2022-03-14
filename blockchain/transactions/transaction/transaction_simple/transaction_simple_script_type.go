package transaction_simple

type ScriptType uint64

const (
	SCRIPT_UPDATE_DELEGATE ScriptType = iota
	SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_UPDATE_DELEGATE:
		return "SCRIPT_UPDATE_DELEGATE"
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		return "SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY"
	default:
		return "Unknown ScriptType"
	}
}
