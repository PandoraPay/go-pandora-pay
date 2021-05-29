package transaction_type

type TransactionType uint64

const (
	TX_SIMPLE TransactionType = iota

	TX_END
)

func (t TransactionType) String() string {
	switch t {
	case TX_SIMPLE:
		return "TX_SIMPLE"
	default:
		return "Unknown transaction type"
	}
}
