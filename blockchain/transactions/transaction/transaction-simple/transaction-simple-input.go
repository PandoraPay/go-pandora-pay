package transaction_simple

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey [33]byte
	Amount    uint64
	Signature [65]byte
	Token     [20]byte

	_publicKeyHash         [20]byte
	_publicKeyHashComputed bool
}

func (vin *TransactionSimpleInput) Serialize(writer *helpers.BufferWriter, inclSignature bool) {
	writer.Write(vin.PublicKey[:])
	writer.WriteUvarint(vin.Amount)
	if inclSignature {
		writer.Write(vin.Signature[:])
	}
	writer.WriteToken(&vin.Token)
}

func (vin *TransactionSimpleInput) Deserialize(reader *helpers.BufferReader) {

	vin.PublicKey = reader.Read33()
	vin.Amount = reader.ReadUvarint()
	vin.Signature = reader.Read65()
	vin.Token = reader.ReadToken()

}

func (vin *TransactionSimpleInput) GetPublicKeyHash() *[20]byte {
	if !vin._publicKeyHashComputed {
		vin._publicKeyHash = *cryptography.ComputePublicKeyHash(&vin.PublicKey)
		vin._publicKeyHashComputed = true
	}
	return &vin._publicKeyHash
}
