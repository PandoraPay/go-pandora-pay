package transaction_transparent

import "pandora-pay/helpers"

type TransactionTransparent struct {
	Nonce uint64
	Vin   []*TransactionTransparentInput
	Vout  []*TransactionTransparentOutput
}

func (tx *TransactionTransparent) Serialize(writer *helpers.BufferWriter) {

}

func (tx *TransactionTransparent) Deserialize(reader *helpers.BufferReader) error {

}
