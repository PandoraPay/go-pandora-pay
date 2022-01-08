package transaction_zether

import (
	"errors"
)

type TransactionZetherBloom struct {
	Nonces         [][]byte
	PublicKeyLists [][][]byte
	bloomed        bool
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
func (tx *TransactionZether) BloomNow() (err error) {

	if tx.Bloom != nil {
		return
	}
	tx.Bloom = new(TransactionZetherBloom)

	if err = tx.bloomLists(); err != nil {
		return
	}

	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionZetherBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	return nil
}
