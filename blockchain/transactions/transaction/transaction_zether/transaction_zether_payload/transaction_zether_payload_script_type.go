package transaction_zether_payload

type PayloadScriptType uint64

const (
	SCRIPT_TRANSFER PayloadScriptType = iota
	SCRIPT_STAKING
	SCRIPT_STAKING_REWARD
	SCRIPT_ASSET_CREATE
	SCRIPT_ASSET_SUPPLY_INCREASE
)

func (t PayloadScriptType) String() string {
	switch t {
	case SCRIPT_TRANSFER:
		return "SCRIPT_TRANSFER"
	case SCRIPT_STAKING:
		return "SCRIPT_STAKING"
	case SCRIPT_STAKING_REWARD:
		return "SCRIPT_STAKING_REWARD"
	case SCRIPT_ASSET_CREATE:
		return "SCRIPT_ASSET_CREATE"
	case SCRIPT_ASSET_SUPPLY_INCREASE:
		return "SCRIPT_ASSET_SUPPLY_INCREASE"
	default:
		return "Unknown ScriptType"
	}
}
