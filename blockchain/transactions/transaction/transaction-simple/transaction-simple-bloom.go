package transaction_simple

type TransactionSimpleBloom struct {
	hashForSignature  []byte
	signatureVerified bool
	bloomed           bool
}

func (tx *TransactionSimple) BloomNow(hashForSignature []byte, signatureWasVerifiedBefore bool) {

	bloom := new(TransactionSimpleBloom)
	bloom.hashForSignature = hashForSignature

	for _, vin := range tx.Vin {
		vin.BloomNow(hashForSignature)
	}

	bloom.signatureVerified = signatureWasVerifiedBefore
	if !signatureWasVerifiedBefore {
		bloom.signatureVerified = tx.VerifySignatureManually(hashForSignature)
		if !bloom.signatureVerified {
			panic("Signature Failed for Transaction Simple")
		}
	}
	bloom.bloomed = true
	tx.Bloom = bloom
}

func (tx *TransactionSimpleBloom) verifyIfBloomed() {
	if !tx.bloomed {
		panic("TransactionSimpleBloom was not bloomed")
	}
}
