package transaction_zether

import (
	"bytes"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZether struct {
	transaction_base_interface.TransactionBaseInterface
	Height         uint64
	Registrations  *transaction_zether_registrations.TransactionZetherDataRegistrations
	Payloads       []*transaction_zether_payload.TransactionZetherPayload
	Bloom          *TransactionZetherBloom
	publicKeysList [][]byte //it calculated with
}

func (tx *TransactionZether) ComputeExtraSpace() uint64 {
	return uint64(64 * len(tx.Registrations.Registrations))
}

/**
Zether requires another verification that the bloomed publicKeys, CL, CR are the same
*/
func (tx *TransactionZether) IncludeTransaction(blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	if err = tx.Registrations.RegisterNow(dataStorage, tx.Bloom.publicKeyListByCounter); err != nil {
		return
	}

	if tx.Height > blockHeight {
		return fmt.Errorf("Zether TxHeight is invalid %d > %d", tx.Height, blockHeight)
	}

	counter := 0
	for payloadIndex, payload := range tx.Payloads {
		if err = payload.IncludePayload(tx.Registrations, payloadIndex, tx.Bloom.publicKeyListByCounter, blockHeight, dataStorage, &counter); err != nil {
			return
		}
	}

	return nil
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
		if err = payload.Validate(tx.Registrations, payloadIndex); err != nil {
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

	tx.Registrations.Serialize(w)

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

	if tx.Height, err = r.ReadUvarint(); err != nil {
		return
	}

	tx.Registrations = new(transaction_zether_registrations.TransactionZetherDataRegistrations)
	if err = tx.Registrations.Deserialize(r); err != nil {
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

	return
}

func (tx *TransactionZether) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
