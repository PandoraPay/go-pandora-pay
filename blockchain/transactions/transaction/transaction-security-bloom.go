package transaction

type TransactionSecurityBloom struct {
	SignatureVerified bool
	bloomed           bool
}

func (tx *Transaction) BloomSecurityNow() {
	tx.BloomSecurity = new(TransactionSecurityBloom)
	tx.BloomSecurity.BloomTransactionSecurity(tx)
}

func (bloom *TransactionSecurityBloom) BloomTransactionSecurity(tx *Transaction) {
	tx.Bloom.VerifyIfBloomed()
	bloom.SignatureVerified = tx.VerifySignature()
	bloom.bloomed = true
}

func (bloom *TransactionSecurityBloom) VerifyIfBloomed() {
	if !bloom.bloomed {
		panic("Tx Security is not bloomed")
	}
}
