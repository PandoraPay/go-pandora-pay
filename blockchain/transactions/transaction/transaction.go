package transaction

import (
	"errors"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Transaction struct {
	transaction_base_interface.TransactionBaseInterface
	Version     transaction_type.TransactionVersion
	DataVersion transaction_type.TransactionDataVersion
	Data        []byte
	Bloom       *TransactionBloom
}

func (tx *Transaction) GetAllFees() (map[string]uint64, error) {
	fees := make(map[string]uint64)
	return fees, tx.ComputeFees(fees)
}

func (tx *Transaction) GetAllKeys() map[string]bool {
	out := make(map[string]bool)
	tx.ComputeAllKeys(out)
	return out
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

	writer.WriteUvarint(uint64(tx.Version))

	writer.WriteByte(byte(tx.DataVersion))
	if tx.DataVersion != transaction_type.TX_DATA_NONE {
		writer.WriteUvarint(uint64(len(tx.Data)))
		writer.Write(tx.Data)
	}

	tx.TransactionBaseInterface.SerializeAdvanced(writer, inclSignature)
}

func (tx *Transaction) Serialize(writer *helpers.BufferWriter) {
	writer.Write(tx.Bloom.Serialized)
}

func (tx *Transaction) SerializeToBytes() []byte {
	return tx.Bloom.Serialized
}

func (tx *Transaction) SerializeManualToBytes() []byte {
	writer := helpers.NewBufferWriter()
	tx.SerializeAdvanced(writer, true)
	return writer.Bytes()
}

func (tx *Transaction) Validate() error {

	if tx.Version != 0 {
		return errors.New("Version is invalid")
	}
	if tx.Version >= transaction_type.TX_END {
		return errors.New("VersionType is invalid")
	}

	return tx.TransactionBaseInterface.Validate()
}

func (tx *Transaction) Verify() error {
	return tx.VerifyBloomAll()
}

func (tx *Transaction) Deserialize(reader *helpers.BufferReader) (err error) {

	first := reader.Position

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.Version = transaction_type.TransactionVersion(n)

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		tx.TransactionBaseInterface = &transaction_simple.TransactionSimple{}
	default:
		return errors.New("Invalid TxType")
	}

	var dataVersion byte
	if dataVersion, err = reader.ReadByte(); err != nil {
		return
	}

	tx.DataVersion = transaction_type.TransactionDataVersion(dataVersion)
	switch tx.DataVersion {
	case transaction_type.TX_DATA_NONE:
	case transaction_type.TX_DATA_PLAIN_TEXT, transaction_type.TX_DATA_ENCRYPTED:
		if n, err = reader.ReadUvarint(); err != nil {
			return
		}
		if n == 0 || n > config.TRANSACTIONS_MAX_DATA_LENGTH {
			return errors.New("Tx.Data length is invalid")
		}
		if tx.Data, err = reader.ReadBytes(int(n)); err != nil {
			return
		}
	default:
		return errors.New("Invalid Tx.DataVersion")
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
