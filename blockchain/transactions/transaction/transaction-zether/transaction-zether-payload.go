package transaction_zether

import (
	"errors"
	"math"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayload struct {
	Token     []byte
	BurnValue uint64

	ExtraType byte   // its unencrypted  and is by default 0 for almost all txs
	ExtraData []byte // rpc payload encryption depends on RPCType

	// sender position in ring representation in a byte, upto 256 ring
	// 144 byte payload  ( to implement specific functionality such as delivery of keys etc), user dependent encryption
	Statement *crypto.Statement // note statement containts fees
	Proof     *crypto.Proof
}

func (payload *TransactionZetherPayload) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteToken(payload.Token)
	w.WriteUvarint(payload.BurnValue)
	w.WriteByte(payload.ExtraType)
	w.Write(payload.ExtraData)
	payload.Statement.Serialize(w)

	if inclSignature {
		payload.Proof.Serialize(w)
	}

}

func (payload *TransactionZetherPayload) Deserialize(reader *helpers.BufferReader) (err error) {

	if payload.Token, err = reader.ReadToken(); err != nil {
		return
	}
	if payload.BurnValue, err = reader.ReadUvarint(); err != nil {
		return
	}
	if payload.ExtraType, err = reader.ReadByte(); err != nil {
		return
	}
	if payload.ExtraData, err = reader.ReadBytes(PAYLOAD_LIMIT); err != nil {
		return
	}
	if err = payload.Statement.Deserialize(reader); err != nil {
		return
	}

	N := len(payload.Statement.Publickeylist)
	m := int(math.Log2(float64(N)))
	if math.Pow(2, float64(m)) != float64(N) {
		return errors.New("log failed")
	}

	if err = payload.Proof.Deserialize(reader, m); err != nil {
		return
	}

	return
}
