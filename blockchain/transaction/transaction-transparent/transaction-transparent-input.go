package transaction_transparent

type TransactionTransparentInput struct {
	PublicKey [33]byte
	Amount    uint64
	Signature [65]byte
}
