package transaction_simple

type ScriptType uint64

const (
	SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY ScriptType = iota
	SCRIPT_RESOLUTION_PAY_IN_FUTURE
)

func (t ScriptType) String() string {
	switch t {
	case SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		return "SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY"
	case SCRIPT_RESOLUTION_PAY_IN_FUTURE:
		return "SCRIPT_RESOLUTION_PAY_IN_FUTURE"
	default:
		return "Unknown ScriptType"
	}
}
