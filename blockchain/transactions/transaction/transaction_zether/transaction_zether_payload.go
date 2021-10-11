package transaction_zether

import (
	"errors"
	"math"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/config"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayload struct {
	Token     []byte
	BurnValue uint64

	DataVersion transaction_data.TransactionDataVersion
	Data        []byte // sender position in ring representation in a byte, upto 256 ring
	// 144 byte payload  ( to implement specific functionality such as delivery of keys etc), user dependent encryption
	Statement *crypto.Statement // note statement containts fees
	Proof     *crypto.Proof
}

func (payload *TransactionZetherPayload) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteToken(payload.Token)
	w.WriteUvarint(payload.BurnValue)

	w.WriteByte(byte(payload.DataVersion))
	if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT { //variable
		w.WriteUvarint(uint64(len(payload.Data)))
		w.Write(payload.Data)
	} else if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED { //fixed 145
		w.Write(payload.Data)
	}

	payload.Statement.Serialize(w)

	if inclSignature {
		payload.Proof.Serialize(w)
	}

}

func (payload *TransactionZetherPayload) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64

	if payload.Token, err = r.ReadToken(); err != nil {
		return
	}
	if payload.BurnValue, err = r.ReadUvarint(); err != nil {
		return
	}

	var dataVersion byte
	if dataVersion, err = r.ReadByte(); err != nil {
		return
	}

	payload.DataVersion = transaction_data.TransactionDataVersion(dataVersion)

	switch payload.DataVersion {
	case transaction_data.TX_DATA_NONE:
	case transaction_data.TX_DATA_PLAIN_TEXT:
		if n, err = r.ReadUvarint(); err != nil {
			return
		}
		if n == 0 || n > config.TRANSACTIONS_MAX_DATA_LENGTH {
			return errors.New("Tx.Data length is invalid")
		}
		if payload.Data, err = r.ReadBytes(int(n)); err != nil {
			return
		}
	case transaction_data.TX_DATA_ENCRYPTED:
		if payload.Data, err = r.ReadBytes(PAYLOAD_LIMIT); err != nil {
			return
		}
	default:
		return errors.New("Invalid Tx.DataVersion")
	}

	if err = payload.Statement.Deserialize(r); err != nil {
		return
	}

	m := int(math.Log2(float64(payload.Statement.RingSize)))
	if math.Pow(2, float64(m)) != float64(payload.Statement.RingSize) {
		return errors.New("log failed")
	}

	if err = payload.Proof.Deserialize(r, m); err != nil {
		return
	}

	return
}
