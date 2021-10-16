package transaction_zether

import (
	"bytes"
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZether struct {
	transaction_base_interface.TransactionBaseInterface
	Extra    transaction_zether_extra.TransactionZetherExtraInterface
	TxScript ScriptType
	Height   uint64
	Payloads []*transaction_zether_payload.TransactionZetherPayload
	Bloom    *TransactionZetherBloom
}

/**
Zether requires another verification that the bloomed publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts
	var acc *account.Account
	var acckey crypto.Point

	c := 0
	for _, payload := range tx.Payloads {
		c += len(payload.Statement.Publickeylist)
	}

	publicKeyList := make([][]byte, c)
	c = 0
	for _, payload := range tx.Payloads {
		for _, publicKey := range payload.Statement.Publickeylist {
			publicKeyList[c] = publicKey.EncodeCompressed()
			c += 1
		}
	}

	if err = txRegistrations.RegisterNow(dataStorage.Regs, publicKeyList); err != nil {
		return
	}

	c = 0
	for _, payload := range tx.Payloads {
		if accs, err = dataStorage.AccsCollection.GetMap(payload.Asset); err != nil {
			return
		}

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

			//verify
			if payload.Statement.CLn[i].String() != balance.Left.String() || payload.Statement.CRn[i].String() != balance.Right.String() {
				return errors.New("CLn or CRn is not matching")
			}

			if acc == nil {
				if acc, err = accs.CreateAccount(publicKey); err != nil {
					return
				}
			}

			acc.Balance.Amount = balance
			if err = accs.Update(string(publicKey), acc); err != nil {
				return
			}
		}
	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	case SCRIPT_DELEGATE_STAKE:
		if err = tx.Extra.IncludeTransaction(txRegistrations, tx.Payloads, blockHeight, dataStorage); err != nil {
			return
		}
	default:
		return errors.New("Invalid tx.TxScript")
	}

	return nil
}

func (tx *TransactionZether) ComputeFees() (uint64, error) {

	sum := uint64(0)
	for _, payload := range tx.Payloads {
		if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET) {
			if err := helpers.SafeUint64Add(&sum, payload.Statement.Fees); err != nil {
				return 0, err
			}
		}
	}

	return sum, nil
}

func (tx *TransactionZether) ComputeAllKeys(out map[string]bool) {
	for _, payload := range tx.Payloads {
		for _, publicKey := range payload.Statement.Publickeylist {
			out[string(publicKey.EncodeCompressed())] = true
		}
	}
	switch tx.TxScript {
	case SCRIPT_DELEGATE_STAKE:
		extra := tx.Extra.(*transaction_zether_extra.TransactionZetherDelegateStake)
		out[string(extra.DelegatePublicKey)] = true
	}

	return
}

func (tx *TransactionZether) Validate() (err error) {

	for _, payload := range tx.Payloads {

		if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
			return errors.New("NATIVE_ASSET_FULL should be written as NATIVE_ASSET")
		}

		// check sanity
		if payload.Statement.RingSize < 2 { // ring size minimum 4
			return fmt.Errorf("RingSize cannot be less than 2")
		}

		if payload.Statement.RingSize >= config.TRANSACTIONS_ZETHER_RING_MAX { // ring size current limited to 256
			return fmt.Errorf("RingSize cannot be that big")
		}

		if !crypto.IsPowerOf2(int(payload.Statement.RingSize)) {
			return fmt.Errorf("corrupted key pointers")
		}

		// check duplicate ring members within the tx
		key_map := map[string]bool{}
		for i := 0; i < int(payload.Statement.RingSize); i++ {
			key_map[string(payload.Statement.Publickeylist[i].EncodeCompressed())] = true
		}
		if len(key_map) != int(payload.Statement.RingSize) {
			return fmt.Errorf("Duplicated ring members")
		}

	}

	switch tx.TxScript {
	case SCRIPT_TRANSFER:
	case SCRIPT_DELEGATE_STAKE, SCRIPT_CLAIM_STAKE:
		if tx.Extra == nil {
			return errors.New("extra is not assigned")
		}
		if err = tx.Extra.Validate(tx.Payloads); err != nil {
			return
		}
	default:
		return errors.New("Invalid Zether TxScript")
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

	w.WriteUvarint(uint64(len(tx.Payloads)))
	for _, payload := range tx.Payloads {
		payload.Serialize(w, inclSignature)
	}

	if tx.Extra != nil {
		tx.Extra.Serialize(w, inclSignature)
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

	tx.TxScript = ScriptType(n)
	switch tx.TxScript {
	case SCRIPT_TRANSFER:
		tx.Extra = nil
	case SCRIPT_DELEGATE_STAKE:
		tx.Extra = &transaction_zether_extra.TransactionZetherDelegateStake{}
	case SCRIPT_CLAIM_STAKE:
		tx.Extra = &transaction_zether_extra.TransactionZetherClaimStake{}
	default:
		return errors.New("INVALID SCRIPT TYPE")
	}

	if tx.Height, err = r.ReadUvarint(); err != nil {
		return
	}

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	for i := uint64(0); i < n; i++ {
		payload := transaction_zether_payload.TransactionZetherPayload{
			Statement: &crypto.Statement{},
			Proof:     &crypto.Proof{},
		}
		if err = payload.Deserialize(r); err != nil {
			return
		}
		tx.Payloads = append(tx.Payloads, &payload)
	}

	if tx.Extra != nil {
		return tx.Extra.Deserialize(r)
	}

	return
}

func (tx *TransactionZether) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
