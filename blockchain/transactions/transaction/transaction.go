package transaction

import (
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Transaction struct {
	helpers.SerializableInterface
	Version uint64
	TxType  transaction_type.TransactionType
	TxBase  transaction_base_interface.TransactionBaseInterface
	Bloom   *TransactionBloom `json:"-"`
}

func (tx *Transaction) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) error {
	return tx.TxBase.IncludeTransaction(blockHeight, accs, toks)
}

func (tx *Transaction) AddFees(fees map[string]uint64) error {
	return tx.TxBase.ComputeFees(fees)
}

func (tx *Transaction) ComputeFees() (fees map[string]uint64, err error) {
	fees = make(map[string]uint64)
	err = tx.AddFees(fees)
	return
}

func (tx *Transaction) SerializeForSigning() []byte {
	writer := helpers.NewBufferWriter()
	tx.SerializeAdvanced(writer, false)
	return cryptography.SHA3Hash(writer.Bytes())
}

func (tx *Transaction) VerifySignatureManually() bool {
	hash := tx.SerializeForSigning()
	return tx.TxBase.VerifySignatureManually(hash)
}

func (tx *Transaction) computeHash() []byte {
	return cryptography.SHA3Hash(tx.SerializeToBytes())
}

func (tx *Transaction) SerializeAdvanced(writer *helpers.BufferWriter, inclSignature bool) {

	writer.WriteUvarint(tx.Version)
	writer.WriteUvarint(uint64(tx.TxType))

	tx.TxBase.SerializeAdvanced(writer, inclSignature)
}

func (tx *Transaction) Serialize(writer *helpers.BufferWriter) {
	tx.SerializeAdvanced(writer, true)
}

func (tx *Transaction) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	tx.Serialize(writer)
	return writer.Bytes()
}

func (tx *Transaction) Validate() error {

	if tx.Version != 0 {
		return errors.New("Version is invalid")
	}
	if transaction_type.TxEND <= tx.TxType {
		return errors.New("VersionType is invalid")
	}

	return tx.TxBase.Validate()
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
	case transaction_type.TxSimple:
		tx.TxBase = &transaction_simple.TransactionSimple{}
	default:
		return errors.New("Invalid TxType")
	}

	if err = tx.TxBase.Deserialize(reader); err != nil {
		return
	}

	end := reader.Position

	//we can bloom more efficiently if asked
	serialized := reader.Buf[first:end]
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
