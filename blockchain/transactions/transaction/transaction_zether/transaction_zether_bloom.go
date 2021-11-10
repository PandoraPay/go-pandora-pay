package transaction_zether

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
)

type TransactionZetherBloom struct {
	Nonces                [][]byte
	PublicKeyLists        [][][]byte
	registrationsVerified bool
	signatureVerified     bool
	bloomed               bool
}

func (tx *TransactionZether) bloomLists() (err error) {
	tx.Bloom.Nonces = make([][]byte, len(tx.Payloads))
	tx.Bloom.PublicKeyLists = make([][][]byte, len(tx.Payloads))
	for t, payload := range tx.Payloads {
		tx.Bloom.PublicKeyLists[t] = make([][]byte, len(payload.Statement.Publickeylist))
		for i := range payload.Statement.Publickeylist {
			tx.Bloom.PublicKeyLists[t][i] = payload.Statement.Publickeylist[i].EncodeCompressed()
		}
		if err = payload.Registrations.ValidateRegistrations(payload.Statement.Publickeylist); err != nil {
			return
		}

		tx.Bloom.Nonces[t] = payload.Proof.Nonce()
	}
	return
}

/**
It blooms publicKeys, CL, CR
*/
func (tx *TransactionZether) BloomNow(hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	if err = tx.bloomLists(); err != nil {
		return
	}

	//verify signature
	assetMap := map[string]int{}
	for _, payload := range tx.Payloads {
		if payload.Proof.Verify(payload.Asset, assetMap[string(payload.Asset)], tx.ChainHash, payload.Statement, hashForSignature, payload.BurnValue) == false {
			return errors.New("Zether Failed for Transaction")
		}
		assetMap[string(payload.Asset)] = assetMap[string(payload.Asset)] + 1
	}

	for _, payload := range tx.Payloads {
		switch payload.PayloadScript {
		case transaction_zether_payload.SCRIPT_DELEGATE_STAKE, transaction_zether_payload.SCRIPT_CLAIM,
			transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
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

	if err = tx.bloomLists(); err != nil {
		return
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
