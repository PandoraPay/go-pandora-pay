package transaction

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_base_interface"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Transaction struct {
	transaction_base_interface.TransactionBaseInterface
	Version transaction_type.TransactionVersion
	Bloom   *TransactionBloom
}

func (tx *Transaction) ComputeExtraSpace() uint64 {
	return tx.TransactionBaseInterface.ComputeExtraSpace()
}

func (tx *Transaction) GetAllFees() (uint64, error) {
	return tx.ComputeFees()
}

func (tx *Transaction) GetAllKeys() map[string]bool {
	out := make(map[string]bool)
	tx.ComputeAllKeys(out)
	return out
}

func (tx *Transaction) SerializeForSigning() []byte {
	writer := helpers.NewBufferWriter()
	tx.SerializeAdvanced(writer, false)
	return cryptography.SHA3(writer.Bytes())
}

func (tx *Transaction) VerifySignatureManually() bool {
	hash := tx.SerializeForSigning()
	return tx.TransactionBaseInterface.VerifySignatureManually(hash)
}

func (tx *Transaction) GetHashSigningManually() []byte {
	return tx.SerializeForSigning()
}

func (tx *Transaction) SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(uint64(tx.Version))

	tx.TransactionBaseInterface.SerializeAdvanced(w, inclSignature)
}

func (tx *Transaction) Serialize(w *helpers.BufferWriter) {
	w.Write(tx.Bloom.Serialized)
}

func (tx *Transaction) SerializeToBytes() []byte {
	return tx.Bloom.Serialized
}

func (tx *Transaction) SerializeManualToBytes() []byte {
	writer := helpers.NewBufferWriter()
	tx.SerializeAdvanced(writer, true)
	return writer.Bytes()
}

func (tx *Transaction) HashManual() []byte {
	serialized := tx.SerializeManualToBytes()
	return cryptography.SHA3(serialized)
}

func (tx *Transaction) validate() error {
	if tx.Version >= transaction_type.TX_END {
		return errors.New("VersionType is invalid")
	}
	return tx.TransactionBaseInterface.Validate()
}

func (tx *Transaction) Verify() error {
	return tx.VerifyBloomAll()
}

func (tx *Transaction) Deserialize(r *helpers.BufferReader) (err error) {

	first := r.Position

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	tx.Version = transaction_type.TransactionVersion(n)

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		tx.TransactionBaseInterface = &transaction_simple.TransactionSimple{}
	case transaction_type.TX_ZETHER:
		tx.TransactionBaseInterface = &transaction_zether.TransactionZether{}
	default:
		return errors.New("Invalid TxType")
	}

	if err = tx.TransactionBaseInterface.Deserialize(r); err != nil {
		return
	}

	//we can bloom more efficiently if asked
	serialized := r.Buf[first:r.Position]
	hash := cryptography.SHA3(serialized)
	tx.Bloom = &TransactionBloom{
		Serialized: serialized,
		Size:       uint64(len(serialized)),
		Hash:       hash,
		HashStr:    string(hash),
		bloomed:    true,
	}

	return
}
