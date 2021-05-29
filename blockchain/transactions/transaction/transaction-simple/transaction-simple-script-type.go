package transaction_simple

type ScriptType uint64

const (
	ScriptNormal ScriptType = iota
	ScriptUnstake
	ScriptWithdraw
	ScriptDelegate

	ScriptEND
)

func (t ScriptType) String() string {
	switch t {
	case ScriptNormal:
		return "ScriptNormal"
	case ScriptUnstake:
		return "ScriptUnstake"
	case ScriptWithdraw:
		return "ScriptWithdraw"
	case ScriptDelegate:
		return "ScriptDelegate"
	default:
		return "Unknown ScriptType"
	}
}
