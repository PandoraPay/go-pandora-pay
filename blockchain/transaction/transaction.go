package transaction

import (
	"errors"
	"pandora-pay/helpers"
)

type TransactionType uint64

const (
	TransactionTransparent TransactionType = 0
)

func (t TransactionType) String() string {
	switch t {
	case TransactionTransparent:
		return "TransactionTransparent"
	default:
		return "Unknown transaction type"
	}
}

type Transaction struct {
	Version         uint64
	TransactionType TransactionType
	TransactionBase interface{}
}

func (tx *Transaction) Serialize() []byte {
	writer := helpers.NewBufferWriter()

	writer.WriteUint64(tx.Version)
	writer.WriteUint64(uint64(tx.TransactionType))

	return writer.Bytes()
}

func (tx *Transaction) Deserialize(buf []byte) (err error) {
	reader := helpers.NewBufferReader(buf)

	if tx.Version, err = reader.ReadUvarint(); err != nil {
		return
	}
	if tx.Version != 0 {
		err = errors.New("Version is invalid")
		return
	}

	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.TransactionType = TransactionType(n)
	if tx.TransactionType != TransactionTransparent {
		errors.New("Transaction Type is invalid")
		return
	}

	return
}
