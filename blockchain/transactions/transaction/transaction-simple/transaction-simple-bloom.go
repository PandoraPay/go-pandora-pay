package transaction_simple

type TransactionSimpleBloom struct {
	hashForSignature  []byte
	signatureVerified bool
	bloomed           bool
}

func (tx *TransactionSimple) BloomNow(hashForSignature []byte, signatureWasVerifiedBefore bool) {

	for _, vin := range tx.Vin {
		vin.BloomNow(hashForSignature)
	}

	tx.Bloom = new(TransactionSimpleBloom)
	tx.Bloom.hashForSignature = hashForSignature
	tx.Bloom.signatureVerified = signatureWasVerifiedBefore
	if !signatureWasVerifiedBefore {
		tx.Bloom.signatureVerified = tx.VerifySignature(hashForSignature)
		if !tx.Bloom.signatureVerified {
			panic("Signature Failed for Transaction Simple")
		}
	}
	tx.Bloom.bloomed = true
}

func (tx *TransactionSimpleBloom) VerifyIfBloomed() {
	if !tx.bloomed {
		panic("TransactionSimpleBloom was not bloomed")
	}
}
