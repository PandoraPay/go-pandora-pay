package transaction

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type Transaction struct {
	Version       uint64
	TxType        transaction_type.TransactionType
	TxBase        interface{}
	Bloom         *TransactionBloom
	BloomSecurity *TransactionSecurityBloom
}

func (tx *Transaction) IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) {
	switch tx.TxType {
	case transaction_type.TxSimple:
		tx.TxBase.(*transaction_simple.TransactionSimple).IncludeTransaction(blockHeight, accs, toks)
	}
}

func (tx *Transaction) AddFees(fees map[string]uint64) {
	switch tx.TxType {
	case transaction_type.TxSimple:
		tx.TxBase.(*transaction_simple.TransactionSimple).ComputeFees(fees)
	}
	return
}

func (tx *Transaction) ComputeFees() (fees map[string]uint64) {
	fees = make(map[string]uint64)
	tx.AddFees(fees)
	return
}

func (tx *Transaction) SerializeForSigning() []byte {
	return cryptography.SHA3Hash(tx.serializeTx(false))
}

func (tx *Transaction) VerifySignature() bool {

	switch tx.TxType {
	case transaction_type.TxSimple:
		return tx.TxBase.(*transaction_simple.TransactionSimple).VerifySignature(tx.Bloom.Hash)
	default:
		return false
	}

}

func (tx *Transaction) ComputeHash() []byte {
	return cryptography.SHA3Hash(tx.Serialize())
}

func (tx *Transaction) serializeTx(inclSignature bool) []byte {

	writer := helpers.NewBufferWriter()

	writer.WriteUvarint(tx.Version)
	writer.WriteUvarint(uint64(tx.TxType))

	switch tx.TxType {
	case transaction_type.TxSimple:
		tx.TxBase.(*transaction_simple.TransactionSimple).Serialize(writer, inclSignature)
	}

	return writer.Bytes()
}

func (tx *Transaction) Serialize() []byte {
	return tx.serializeTx(true)
}

func (tx *Transaction) Validate() {
	if tx.Version != 0 {
		panic("Version is invalid")
	}
	if transaction_type.TxEND < tx.TxType {
		panic("VersionType is invalid")
	}

	switch tx.TxType {
	case transaction_type.TxSimple:
		tx.TxBase.(*transaction_simple.TransactionSimple).Validate()
	}

	return
}

func (tx *Transaction) Verify() {
	tx.VerifyBloomAll() //it will panic
	if tx.BloomSecurity.SignatureVerified != true {
		panic("Signature is invalid")
	}
}

func (tx *Transaction) Deserialize(reader *helpers.BufferReader) {

	tx.Version = reader.ReadUvarint()
	n := reader.ReadUvarint()
	tx.TxType = transaction_type.TransactionType(n)

	switch tx.TxType {
	case transaction_type.TxSimple:
		base := &transaction_simple.TransactionSimple{}
		base.Deserialize(reader)
		tx.TxBase = base
	}

}

func (tx *Transaction) BloomAll() {
	tx.BloomNow()
	tx.BloomSecurityNow()
	switch tx.TxType {
	case transaction_type.TxSimple:
		base := tx.TxBase.(*transaction_simple.TransactionSimple)
		for _, vin := range base.Vin {
			vin.BloomNow(tx.Bloom.Hash)
		}
	}
}

func (tx *Transaction) VerifyBloomAll() {
	tx.Bloom.VerifyIfBloomed()
	tx.BloomSecurity.VerifyIfBloomed()
	switch tx.TxType {
	case transaction_type.TxSimple:
		base := tx.TxBase.(*transaction_simple.TransactionSimple)
		for _, vin := range base.Vin {
			vin.Bloom.VerifyIfBloomed()
		}
	}
}
