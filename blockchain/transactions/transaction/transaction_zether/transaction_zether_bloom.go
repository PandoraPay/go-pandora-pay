package transaction_zether

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_extra"
)

type TransactionZetherBloom struct {
	Nonce1            []byte
	Nonce2            []byte
	signatureVerified bool
	bloomed           bool
}

/**
It blooms publicKeys, CL, CR
*/
func (tx *TransactionZether) BloomNow(hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	//verify signature
	for _, payload := range tx.Payloads {
		if payload.Proof.Verify(payload.Statement, hashForSignature, tx.Height, payload.BurnValue) == false {
			return errors.New("Zether Failed for Transaction")
		}
	}

	tx.Bloom.Nonce1 = tx.Payloads[0].Proof.Nonce1()
	tx.Bloom.Nonce2 = tx.Payloads[0].Proof.Nonce2()

	switch tx.TxScript {
	case SCRIPT_DELEGATE_STAKE:
		extra := tx.Extra.(*transaction_zether_extra.TransactionZetherDelegateStake)
		if extra.DelegatedStakingNewInfo && extra.VerifySignatureManually(hashForSignature) == false {
			return errors.New("DelegatedPublicKey signature failed")
		}
	}

	tx.Bloom.signatureVerified = true
	tx.Bloom.bloomed = true

	return
}

func (tx *TransactionZether) BloomNowSignatureVerified() (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	tx.Bloom.Nonce1 = tx.Payloads[0].Proof.Nonce1()
	tx.Bloom.Nonce2 = tx.Payloads[0].Proof.Nonce2()

	tx.Bloom.signatureVerified = true
	tx.Bloom.bloomed = true

	return
}

func (tx *TransactionZetherBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	if !tx.signatureVerified {
		return errors.New("signatureVerified is false")
	}
	return nil
}
