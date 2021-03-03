package transaction_type

type TransactionType uint64

const (
	TransactionTypeSimple TransactionType = iota
	TransactionTypeSimpleUnstake

	TransactionTypeEND
)

func (t TransactionType) String() string {
	switch t {
	case TransactionTypeSimple:
		return "TransactionSimple"
	case TransactionTypeSimpleUnstake:
		return "TransactionTypeSimpleUnstake"
	default:
		return "Unknown transaction type"
	}
}
