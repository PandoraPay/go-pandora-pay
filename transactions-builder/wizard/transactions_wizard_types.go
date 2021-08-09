package wizard

type TransactionsWizardFee struct {
	Fixed, PerByte uint64
	PerByteAuto    bool
	Token          []byte
}

type TransactionsWizardFeeExtra struct {
	TransactionsWizardFee
	PayInExtra bool
}

type TransactionsWizardData struct {
	Data    []byte
	Encrypt bool
}
