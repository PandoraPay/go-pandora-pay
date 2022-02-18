package transaction_zether

import (
	"bytes"
	"errors"
	"pandora-pay/config/config_coins"
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
			publicKey := payload.Statement.Publickeylist[i].EncodeCompressed()
			if bytes.Equal(publicKey, config_coins.BURN_PUBLIC_KEY) {
				return errors.New("Ring member can not be BURN address")
			}
			tx.Bloom.PublicKeyLists[t][i] = publicKey
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
