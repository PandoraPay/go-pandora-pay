package transaction_type

type TransactionVersion uint64

const (
	TX_SIMPLE TransactionVersion = iota
	TX_ZETHER
	TX_END
)

func (t TransactionVersion) String() string {
	switch t {
	case TX_SIMPLE:
		return "TX_SIMPLE"
	case TX_ZETHER:
		return "TX_ZETHER"
	default:
		return "Unknown transaction type"
	}
}
