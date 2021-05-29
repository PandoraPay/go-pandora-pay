package transaction_type

type TransactionType uint64

const (
	TxSimple TransactionType = iota

	TxEND
)

func (t TransactionType) String() string {
	switch t {
	case TxSimple:
		return "TxSimple"
	default:
		return "Unknown transaction type"
	}
}
