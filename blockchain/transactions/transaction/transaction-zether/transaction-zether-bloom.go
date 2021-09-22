package transaction_zether

import (
	"errors"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
)

type TransactionZetherBloom struct {
	signatureVerified bool
	bloomed           bool
}

/**
Zether requires another verification that the bloomed publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) VerifyBloomNow(registrations *registrations.Registrations, accsCollection *accounts.AccountsCollection, hashForSignature []byte) (err error) {

	for _, payload := range tx.Payloads {

		var accs *accounts.Accounts
		if accs, err = accsCollection.GetMap(payload.Token); err != nil {
			return
		}

		for i, statementPublicKeyPoint := range payload.Statement.Publickeylist {

			publicKey := statementPublicKeyPoint.EncodeCompressed()

			var acc *account.Account
			if acc, err = accs.GetAccount(publicKey); err != nil {
				return
			}

			var a, b *bn256.G1
			if acc == nil {
				var acckey crypto.Point
				if err = acckey.DecodeCompressed(publicKey); err != nil {
					return
				}
				point := crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
				a = point.Left
				b = point.Right
			} else {
				a = acc.Balance.Amount.Left
				b = acc.Balance.Amount.Right
			}

			if payload.Statement.CLn[i].String() != a.String() || payload.Statement.CRn[i].String() != b.String() {
				return errors.New("CLn or CRn is not matching")
			}

		}

	}

	return
}

/**
It blooms publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) BloomNow(hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	for _, payload := range tx.Payloads {
		if payload.Proof.Verify(payload.Statement, hashForSignature, tx.Height, payload.BurnValue) == false {
			return errors.New("Zether Failed for Transaction")
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
