package transaction_zether

import (
	"errors"
	transaction_base_interface "pandora-pay/blockchain/transactions/transaction/transaction-base-interface"
	"pandora-pay/helpers"
)

type TransactionZether struct {
	transaction_base_interface.TransactionBaseInterface
	TxScript ScriptType
	Height   uint64
	Payloads []*TransactionZetherPayload
	Bloom    *TransactionZetherBloom
}

func (tx *TransactionZether) ComputeFees() uint64 {

	//sum := uint64(0)
	//for _, payload := range tx.Payloads{
	//	if err = helpers.SafeUint64Add(&sum, payload.Statement.Fees ); err != nil {
	//		return
	//	}
	//}

	return 0
}

func (tx *TransactionZether) ComputeAllKeys(out map[string]bool) {
	for _, payload := range tx.Payloads {
		for _, publicKey := range payload.Statement.Publickeylist {
			out[string(publicKey.EncodeCompressed())] = true
		}
	}
	return
}

func (tx *TransactionZether) SerializeAdvanced(writer *helpers.BufferWriter, inclSignature bool) {
	writer.WriteUvarint(uint64(tx.TxScript))
	writer.WriteUvarint(tx.Height)
}

func (tx *TransactionZether) Serialize(writer *helpers.BufferWriter) {
	tx.SerializeAdvanced(writer, true)
}

func (tx *TransactionZether) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	tx.Serialize(writer)
	return writer.Bytes()
}

func (tx *TransactionZether) Deserialize(reader *helpers.BufferReader) (err error) {
	var n uint64

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}

	scriptType := ScriptType(n)
	if scriptType >= SCRIPT_END {
		return errors.New("INVALID SCRIPT TYPE")
	}

	if tx.Height, err = reader.ReadUvarint(); err != nil {
		return
	}

	return
}

func (tx *TransactionZether) VerifyBloomAll() (err error) {
	return tx.Bloom.verifyIfBloomed()
}
