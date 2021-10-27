package transaction_zether

import (
	"bytes"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZether struct {
	transaction_base_interface.TransactionBaseInterface
	Height   uint64
	Payloads []*transaction_zether_payload.TransactionZetherPayload
	Bloom    *TransactionZetherBloom
}

func (tx *TransactionZether) ComputeExtraSpace() uint64 {

	totalRegistrations := 0
	for _, payload := range tx.Payloads {
		totalRegistrations += len(payload.Registrations.Registrations)
	}
	return uint64(64 * totalRegistrations)
}

/**
Zether requires another verification that the bloomed publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) IncludeTransaction(blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	if tx.Height > blockHeight {
		return fmt.Errorf("Zether TxHeight is invalid %d > %d", tx.Height, blockHeight)
	}

	for payloadIndex, payload := range tx.Payloads {
		if err = payload.IncludePayload(byte(payloadIndex), tx.Bloom.publicKeyLists[payloadIndex], blockHeight, dataStorage); err != nil {
			return
		}
	}

	return
}

func (tx *TransactionZether) ComputeFees() (uint64, error) {

	sum := uint64(0)
	for _, payload := range tx.Payloads {
		if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
			if err := helpers.SafeUint64Add(&sum, payload.Statement.Fees); err != nil {
				return 0, err
			}
		}
	}

	return sum, nil
}

func (tx *TransactionZether) ComputeAllKeys(out map[string]bool) {

	for _, payload := range tx.Payloads {
		payload.ComputeAllKeys(out)
	}

	return
}

func (tx *TransactionZether) Validate() (err error) {

	for payloadIndex, payload := range tx.Payloads {
		if err = payload.Validate(byte(payloadIndex)); err != nil {
			return
		}
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
	w.WriteUvarint(tx.Height)

	w.WriteByte(byte(len(tx.Payloads)))
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

	if tx.Height, err = r.ReadUvarint(); err != nil {
		return
	}

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}

	tx.Payloads = make([]*transaction_zether_payload.TransactionZetherPayload, n)
	for i := byte(0); i < n; i++ {
		payload := &transaction_zether_payload.TransactionZetherPayload{
			Statement: &crypto.Statement{},
			Proof:     &crypto.Proof{},
		}
		if err = payload.Deserialize(r); err != nil {
			return
		}
		tx.Payloads[i] = payload
	}

	return
}

func (tx *TransactionZether) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
