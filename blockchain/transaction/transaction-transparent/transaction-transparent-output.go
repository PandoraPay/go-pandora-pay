package transaction_transparent

type TransactionTransparentOutput struct {
	PublicKeyHash [20]byte
	Amount        uint64
	Token         [20]byte
}
