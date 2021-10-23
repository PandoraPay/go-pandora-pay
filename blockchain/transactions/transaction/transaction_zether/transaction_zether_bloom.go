package transaction_zether

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
)

type TransactionZetherBloom struct {
	Nonce1                []byte
	Nonce2                []byte
	publicKeyLists        [][][]byte
	registrationsVerified bool
	signatureVerified     bool
	bloomed               bool
}

/**
It blooms publicKeys, CL, CR
*/
func (tx *TransactionZether) BloomNow(hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	tx.Bloom.publicKeyLists = make([][][]byte, len(tx.Payloads))
	for payloadIndex, payload := range tx.Payloads {
		tx.Bloom.publicKeyLists[payloadIndex] = make([][]byte, len(payload.Statement.Publickeylist))
		for i, publicKey := range payload.Statement.Publickeylist {
			tx.Bloom.publicKeyLists[payloadIndex][i] = publicKey.EncodeCompressed()
		}
		if err = payload.Registrations.ValidateRegistrations(payload.Statement.Publickeylist); err != nil {
			return
		}
	}

	//verify signature
	for _, payload := range tx.Payloads {
		if payload.Proof.Verify(payload.Statement, hashForSignature, tx.Height, payload.BurnValue) == false {
			return errors.New("Zether Failed for Transaction")
		}
	}

	tx.Bloom.Nonce1 = tx.Payloads[0].Proof.Nonce1()
	tx.Bloom.Nonce2 = tx.Payloads[0].Proof.Nonce2()

	for _, payload := range tx.Payloads {
		switch payload.PayloadScript {
		case transaction_zether_payload.SCRIPT_DELEGATE_STAKE, transaction_zether_payload.SCRIPT_CLAIM_STAKE:
			if payload.Extra.VerifyExtraSignature(hashForSignature) == false {
				return errors.New("DelegatedPublicKey signature failed")
			}
		}
	}

	tx.Bloom.signatureVerified = true
	tx.Bloom.registrationsVerified = true
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

	c := 0
	for _, payload := range tx.Payloads {
		c += len(payload.Statement.Publickeylist)
	}

	tx.Bloom.publicKeyLists = make([][][]byte, len(tx.Payloads))
	for payloadIndex, payload := range tx.Payloads {
		tx.Bloom.publicKeyLists[payloadIndex] = make([][]byte, len(payload.Statement.Publickeylist))
		for i, publicKey := range payload.Statement.Publickeylist {
			tx.Bloom.publicKeyLists[payloadIndex][i] = publicKey.EncodeCompressed()
		}
	}

	tx.Bloom.signatureVerified = true
	tx.Bloom.registrationsVerified = true
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
	if !tx.registrationsVerified {
		return errors.New("registrationsVerified is false")
	}
	return nil
}
