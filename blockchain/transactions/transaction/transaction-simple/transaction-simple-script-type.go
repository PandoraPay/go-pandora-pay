package transaction_simple

type TransactionSimpleScriptType uint64

const (
	TxSimpleScriptNormal TransactionSimpleScriptType = iota
	TxSimpleScriptUnstake
	TxSimpleScriptWithdraw
	TxSimpleScriptDelegate

	TransactionSimpleScriptEND
)

func (t TransactionSimpleScriptType) String() string {
	switch t {
	case TxSimpleScriptNormal:
		return "TxSimpleScriptNormal"
	case TxSimpleScriptUnstake:
		return "TxSimpleScriptUnstake"
	case TxSimpleScriptWithdraw:
		return "TxSimpleScriptWithdraw"
	case TxSimpleScriptDelegate:
		return "TxSimpleScriptDelegate"
	default:
		return "Unknown TransactionSimpleScriptType"
	}
}
