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
	Version uint64
	TxType  transaction_type.TransactionType
	TxBase  transaction_base_interface.TransactionBaseInterface
	Bloom   *TransactionBloom
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
	return cryptography.SHA3Hash(tx.serializeTx(false))
}

func (tx *Transaction) VerifySignatureManually() bool {
	hash := tx.SerializeForSigning()
	return tx.TxBase.VerifySignatureManually(hash)
}

func (tx *Transaction) ComputeHash() []byte {
	return cryptography.SHA3Hash(tx.Serialize())
}

func (tx *Transaction) serializeTx(inclSignature bool) []byte {

	writer := helpers.NewBufferWriter()

	writer.WriteUvarint(tx.Version)
	writer.WriteUvarint(uint64(tx.TxType))

	tx.TxBase.Serialize(writer, inclSignature)

	return writer.Bytes()
}

func (tx *Transaction) Serialize() []byte {
	return tx.serializeTx(true)
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

func (tx *Transaction) Deserialize(reader *helpers.BufferReader, bloom bool) (err error) {

	buffer := reader.Buf[:]
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

	if bloom {
		//we can bloom more efficiently if asked
		serialized := buffer[first:end]
		hash := cryptography.SHA3(serialized)
		tx.Bloom = &TransactionBloom{
			Serialized: serialized,
			Size:       uint64(len(serialized)),
			Hash:       hash,
			HashStr:    string(hash),
			bloomed:    true,
		}
		if err = tx.BloomExtraNow(true); err != nil {
			return
		}
	}
	return
}
