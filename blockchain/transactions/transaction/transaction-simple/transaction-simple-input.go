package transaction_simple

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey helpers.ByteString //33
	Amount    uint64
	Signature helpers.ByteString //65
	Token     helpers.ByteString //20

	_publicKeyHash []byte //20
}

func (vin *TransactionSimpleInput) Serialize(writer *helpers.BufferWriter, inclSignature bool) {
	writer.Write(vin.PublicKey)
	writer.WriteUvarint(vin.Amount)
	if inclSignature {
		writer.Write(vin.Signature)
	}
	writer.WriteToken(vin.Token)
}

func (vin *TransactionSimpleInput) Deserialize(reader *helpers.BufferReader) {

	vin.PublicKey = reader.ReadBytes(33)
	vin.Amount = reader.ReadUvarint()
	vin.Signature = reader.ReadBytes(65)
	vin.Token = reader.ReadToken()

}

func (vin *TransactionSimpleInput) GetPublicKeyHash() []byte {
	if vin._publicKeyHash == nil {
		vin._publicKeyHash = cryptography.ComputePublicKeyHash(vin.PublicKey)
	}
	return vin._publicKeyHash
}
