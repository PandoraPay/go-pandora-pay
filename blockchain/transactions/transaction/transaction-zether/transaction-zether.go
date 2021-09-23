package transaction_zether

import (
	"errors"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZether struct {
	transaction_base_interface.TransactionBaseInterface
	TxScript      ScriptType
	Height        uint64
	Registrations []*TransactionZetherRegistration
	Payloads      []*TransactionZetherPayload
	Bloom         *TransactionZetherBloom
}

/**
Zether requires another verification that the bloomed publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) IncludeTransaction(blockHeight uint64, regs *registrations.Registrations, plainAccs *plain_accounts.PlainAccounts, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) (err error) {

	var accs *accounts.Accounts
	var acc *account.Account
	var acckey crypto.Point

	c := 0
	for _, payload := range tx.Payloads {
		c += len(payload.Statement.Publickeylist)
	}

	publicKeyList := make([][]byte, c)
	c = 0

	//verification that the bloomed publicKeys, CL, CR are the same
	for _, payload := range tx.Payloads {

		if accs, err = accsCollection.GetMap(payload.Token); err != nil {
			return
		}

		for i, statementPublicKeyPoint := range payload.Statement.Publickeylist {

			publicKey := statementPublicKeyPoint.EncodeCompressed()

			publicKeyList[c] = publicKey
			c += 1

			if acc, err = accs.GetAccount(publicKey); err != nil {
				return
			}

			var a, b *bn256.G1
			if acc == nil { //zero balance
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
	//verify that the other accounts did not register meanwhile
	for _, reg := range tx.Registrations {
		var isReg bool
		if isReg, err = regs.Exists(string(publicKeyList[reg.PublicKeyIndex])); err != nil {
			return
		}
		if isReg {
			return errors.New("PublicKey is already registered")
		}
	}

	//let's register
	for _, reg := range tx.Registrations {
		if _, err = regs.CreateRegistration(publicKeyList[reg.PublicKeyIndex], reg.RegistrationSignature); err != nil {
			return
		}
	}

	c = 0
	for _, payload := range tx.Payloads {
		for i := range payload.Statement.Publickeylist {

			publicKey := publicKeyList[c]
			c += 1

			if acc, err = accs.GetAccount(publicKey); err != nil {
				return
			}

			var balance *crypto.ElGamal
			if acc == nil { //zero balance
				if err = acckey.DecodeCompressed(publicKey); err != nil {
					return
				}
				balance = crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
			} else {
				balance = acc.GetBalance()
			}
			echanges := crypto.ConstructElGamal(payload.Statement.C[i], payload.Statement.D)
			balance = balance.Add(echanges) // homomorphic addition of changes

			acc.Balance.Amount = balance
			if err = accs.Update(string(publicKey), acc); err != nil {
				return
			}
		}
	}

	return nil
}

func (tx *TransactionZether) ComputeFees() uint64 {

	//sum := uint64(0)
	//for _, payload := range tx.Payloads{
	//	if err = helpers.SafeUint64Add(&sum, payload.Statement.Fees ); err != nil {
	//		return
	//	}
	//}

	return 0
}

func (tx *TransactionZether) ComputeAllKeys(out map[string]bool) {
	for _, payload := range tx.Payloads {
		for _, publicKey := range payload.Statement.Publickeylist {
			out[string(publicKey.EncodeCompressed())] = true
		}
	}
	return
}

func (tx *TransactionZether) Validate() (err error) {

	c := 0
	for _, payload := range tx.Payloads {
		c += len(payload.Statement.Publickeylist)
	}

	publicKeyList := make([][]byte, c)

	c = 0
	for _, payload := range tx.Payloads {
		for _, publicKeyPoint := range payload.Statement.Publickeylist {
			publicKeyList[c] = publicKeyPoint.EncodeCompressed()
			c += 1
		}
	}

	uniqueMap := make(map[string]bool)
	for _, registration := range tx.Registrations {
		publicKey := publicKeyList[registration.PublicKeyIndex]
		if publicKey == nil {
			return errors.New("Public Key is not set")
		}
		if uniqueMap[string(publicKey)] != false {
			return errors.New("registration.PublicKey exists multiple times")
		}
		uniqueMap[string(publicKey)] = true
	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER, SCRIPT_DELEGATE:
	default:
		return errors.New("Invalid TxScript")
	}

	return
}

func (tx *TransactionZether) VerifySignatureManually(hash []byte) bool {

	for t := range tx.Payloads {
		if tx.Payloads[t].Proof.Verify(tx.Payloads[t].Statement, hash, tx.Height, tx.Payloads[t].BurnValue) == false {
			return false
		}
	}

	return true
}

func (tx *TransactionZether) SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(uint64(tx.TxScript))
	w.WriteUvarint(tx.Height)

	w.WriteUvarint(uint64(len(tx.Registrations)))
	for _, registration := range tx.Registrations {
		registration.Serialize(w)
	}

	w.WriteUvarint(uint64(len(tx.Payloads)))
	for _, payload := range tx.Payloads {
		payload.Serialize(w, inclSignature)
	}
}

func (tx *TransactionZether) Serialize(w *helpers.BufferWriter) {
	tx.SerializeAdvanced(w, true)
}

func (tx *TransactionZether) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	tx.Serialize(w)
	return w.Bytes()
}

func (tx *TransactionZether) Deserialize(r *helpers.BufferReader) (err error) {
	var n uint64

	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	scriptType := ScriptType(n)
	if scriptType >= SCRIPT_END {
		return errors.New("INVALID SCRIPT TYPE")
	}

	if tx.Height, err = r.ReadUvarint(); err != nil {
		return
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	tx.Registrations = make([]*TransactionZetherRegistration, n)
	for i := uint64(0); i < n; i++ {
		registration := &TransactionZetherRegistration{}
		if err = registration.Deserialize(r); err != nil {
			return
		}
		tx.Registrations[i] = registration
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	for i := uint64(0); i < n; i++ {
		payload := TransactionZetherPayload{
			Statement: &crypto.Statement{},
			Proof:     &crypto.Proof{},
		}
		if err = payload.Deserialize(r); err != nil {
			return
		}
		tx.Payloads = append(tx.Payloads, &payload)
	}

	return
}

func (tx *TransactionZether) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
