package transaction_data

type TransactionDataVersion byte

const (
	TX_DATA_NONE TransactionDataVersion = iota
	TX_DATA_PLAIN_TEXT
	TX_DATA_ENCRYPTED
	TX_DATA_END
)

func (t TransactionDataVersion) String() string {
	switch t {
	case TX_DATA_NONE:
		return "TX_DATA_NONE"
	case TX_DATA_PLAIN_TEXT:
		return "TX_DATA_PLAIN_TEXT"
	case TX_DATA_ENCRYPTED:
		return "TX_DATA_ENCRYPTED"
	default:
		return "Unknown transaction data type"
	}
}
