package transaction_simple

type ScriptType uint64

const (
	SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY ScriptType = iota
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		return "SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY"
	default:
		return "Unknown ScriptType"
	}
}
