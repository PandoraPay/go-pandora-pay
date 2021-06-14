package transaction

import (
	"errors"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Transaction struct {
	transaction_base_interface.TransactionBaseInterface `json:"base"`
	Version                                             uint64                           `json:"version"`
	TxType                                              transaction_type.TransactionType `json:"txType"`
	Bloom                                               *TransactionBloom                `json:"bloom"`
}

func (tx *Transaction) GetAllFees() (map[string]uint64, error) {
	fees := make(map[string]uint64)
	return fees, tx.ComputeFees(fees)
}

func (tx *Transaction) GetAllKeys() (map[string]bool, error) {
	out := make(map[string]bool)
	return out, tx.ComputeAllKeys(out)
}

func (tx *Transaction) SerializeForSigning() []byte {
	writer := helpers.NewBufferWriter()
	tx.SerializeAdvanced(writer, false)
	return cryptography.SHA3Hash(writer.Bytes())
}

func (tx *Transaction) VerifySignatureManually() bool {
	hash := tx.SerializeForSigning()
	return tx.TransactionBaseInterface.VerifySignatureManually(hash)
}

func (tx *Transaction) computeHash() []byte {
	return cryptography.SHA3Hash(tx.SerializeToBytes())
}

func (tx *Transaction) SerializeAdvanced(writer *helpers.BufferWriter, inclSignature bool) {

	writer.WriteUvarint(tx.Version)
	writer.WriteUvarint(uint64(tx.TxType))

	tx.TransactionBaseInterface.SerializeAdvanced(writer, inclSignature)
}

func (tx *Transaction) Serialize(writer *helpers.BufferWriter) {
	tx.SerializeAdvanced(writer, true)
}

func (tx *Transaction) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	tx.Serialize(writer)
	return writer.Bytes()
}

func (tx *Transaction) SerializeToBytesBloomed() []byte {
	if tx.Bloom != nil {
		return tx.Bloom.Serialized
	}
	return tx.SerializeToBytes()
}

func (tx *Transaction) Validate() error {

	if tx.Version != 0 {
		return errors.New("Version is invalid")
	}
	if tx.TxType >= transaction_type.TX_END {
		return errors.New("VersionType is invalid")
	}

	return tx.TransactionBaseInterface.Validate()
}

func (tx *Transaction) Verify() error {
	return tx.VerifyBloomAll()
}

func (tx *Transaction) Deserialize(reader *helpers.BufferReader) (err error) {

	first := reader.Position

	if tx.Version, err = reader.ReadUvarint(); err != nil {
		return
	}

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.TxType = transaction_type.TransactionType(n)

	switch tx.TxType {
	case transaction_type.TX_SIMPLE:
		tx.TransactionBaseInterface = &transaction_simple.TransactionSimple{}
	default:
		return errors.New("Invalid TxType")
	}

	if err = tx.TransactionBaseInterface.Deserialize(reader); err != nil {
		return
	}

	//we can bloom more efficiently if asked
	serialized := reader.Buf[first:reader.Position]
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
