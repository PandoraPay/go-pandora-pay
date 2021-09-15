package transaction_zether

import (
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/registrations"
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

		for i, statementPublicKeyIndex := range payload.Statement.PublicKeysIndexes {

			var publicKey []byte
			if statementPublicKeyIndex.Registered {
				if publicKey, err = registrations.GetKeyByIndex(statementPublicKeyIndex.Index); err != nil {
					return
				}
			} else {
				publicKey = tx.Registrations[statementPublicKeyIndex.Index].PublicKey
			}

			var point crypto.Point
			if err = point.DecodeCompressed(publicKey); err != nil {
				return err
			}

			if payload.Statement.Publickeylist[i].String() != point.G1().String() {
				return errors.New("Publickey is not matching")
			}

			var acc *account.Account
			if acc, err = accs.GetAccount(publicKey, tx.Height); err != nil {
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
func (tx *TransactionZether) BloomNow(registrations *registrations.Registrations, accsCollection *accounts.AccountsCollection, hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	uniqueMap := make(map[string]bool)
	for _, registration := range tx.Registrations {
		if uniqueMap[string(registration.PublicKey)] != false {
			return errors.New("registration.PublicKey exists multiple times")
		}
		uniqueMap[string(registration.PublicKey)] = true
	}

	for _, payload := range tx.Payloads {

		var accs *accounts.Accounts
		if accs, err = accsCollection.GetMap(payload.Token); err != nil {
			return
		}

		for i, statementPublicKeyIndex := range payload.Statement.PublicKeysIndexes {

			var publicKey []byte
			if statementPublicKeyIndex.Registered {
				if publicKey, err = registrations.GetKeyByIndex(statementPublicKeyIndex.Index); err != nil {
					return
				}
			} else {
				if statementPublicKeyIndex.Index > uint64(len(tx.Registrations)) {
					return errors.New("statementPublicKeyIndex is pointing to an unregistered public key")
				}
				publicKey = tx.Registrations[statementPublicKeyIndex.Index].PublicKey
			}

			var point crypto.Point
			if err = point.DecodeCompressed(publicKey); err != nil {
				return err
			}
			payload.Statement.Publickeylist[i] = point.G1()

			var acc *account.Account
			if acc, err = accs.GetAccount(publicKey, tx.Height); err != nil {
				return
			}

			if acc == nil {
				var acckey crypto.Point
				if err = acckey.DecodeCompressed(publicKey); err != nil {
					return
				}
				point := crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
				payload.Statement.CLn[i] = point.Left
				payload.Statement.CRn[i] = point.Right
			} else {
				payload.Statement.CLn[i] = acc.Balance.Amount.Left
				payload.Statement.CRn[i] = acc.Balance.Amount.Right
			}

		}

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
