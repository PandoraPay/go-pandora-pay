package transaction_zether

import (
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayload struct {
	helpers.SerializableInterface

	Token     []byte
	BurnValue uint64

	ExtraType byte   // its unencrypted  and is by default 0 for almost all txs
	ExtraData []byte // rpc payload encryption depends on RPCType

	// sender position in ring representation in a byte, upto 256 ring
	// 144 byte payload  ( to implement specific functionality such as delivery of keys etc), user dependent encryption
	Statement *crypto.Statement // note statement containts fees
	Proof     *crypto.Proof
}

func (payload *TransactionZetherPayload) Serialize(writer *helpers.BufferWriter) {
	writer.WriteToken(payload.Token)
}

func (payload *TransactionZetherPayload) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	payload.Serialize(writer)
	return writer.Bytes()
}

func (payload *TransactionZetherPayload) Deserialize(reader *helpers.BufferReader) (err error) {

	if payload.Token, err = reader.ReadToken(); err != nil {
		return
	}

	return
}
