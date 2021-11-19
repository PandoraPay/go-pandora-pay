package transaction_zether_registration

type TransactionZetherDataRegistrationType byte

const (
	NOT_REGISTERED TransactionZetherDataRegistrationType = iota
	REGISTERED_EMPTY_ACCOUNT
	REGISTERED_ACCOUNT
)
