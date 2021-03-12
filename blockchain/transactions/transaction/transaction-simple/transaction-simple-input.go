package transaction_simple

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey []byte //33
	Amount    uint64
	Signature []byte //65
	Token     []byte //20

	_publicKeyHash         []byte //20
	_publicKeyHashComputed bool
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
	if !vin._publicKeyHashComputed {
		vin._publicKeyHash = cryptography.ComputePublicKeyHash(vin.PublicKey)
		vin._publicKeyHashComputed = true
	}
	return vin._publicKeyHash
}
