package transaction_zether

import (
	"errors"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
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
