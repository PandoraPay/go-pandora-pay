package transaction_transparent

import "pandora-pay/helpers"

type TransactionTransparentOutput struct {
	PublicKeyHash [20]byte
	Amount        uint64
	Token         [20]byte
}

func (tx *TransactionTransparentInput) Serialize(writer *helpers.BufferWriter) {

}

func (tx *TransactionTransparentInput) Deserialize(reader *helpers.BufferReader) error {

}
