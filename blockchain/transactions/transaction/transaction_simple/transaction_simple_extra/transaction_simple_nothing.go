package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers/advanced_buffers"
	"strconv"
)

type TransactionSimpleNothing struct {
	TransactionSimpleExtraInterface
}

// func (this *TransactionSimpleNothing) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {


// 	return
// }

func (this *TransactionSimpleNothing) MessageForSigning() []byte {
	w := advanced_buffers.NewBufferWriter()
	w.Write(this.TxId)
	w.WriteByte(this.PayloadIndex)
	w.WriteBool(this.Resolution)
	return cryptography.SHA3(w.Bytes())
}

func (this *TransactionSimpleNothing) VerifySignature() bool {
	for i := range this.MultisigPublicKeys {
		msg := this.MessageForSigning()
		if !crypto.VerifySignature(msg, this.Signatures[i], this.MultisigPublicKeys[i]) {
			return false
		}
	}
	return true
}

func (this *TransactionSimpleNothing) Validate(fee uint64) (err error) {
	if len(this.MultisigPublicKeys) != len(this.Signatures) {
		return errors.New("Signatures and Public Keys Mismatch")
	}
	if len(this.MultisigPublicKeys) == 0 || len(this.MultisigPublicKeys) > 5 {
		return errors.New("Invalid number of Public Keys")
	}
	if fee != 0 {
		return errors.New("Fee should be zero")
	}
	unique := make(map[string]bool)
	for i := range this.MultisigPublicKeys {
		unique[string(this.MultisigPublicKeys[i])] = true
	}
	if len(unique) != len(this.MultisigPublicKeys) {
		return errors.New("public Keys contain duplicates")
	}
	return
}

func (this *TransactionSimpleNothing) Serialize(w *advanced_buffers.BufferWriter, inclSignature bool) {
	w.Write(this.TxId)
	w.WriteByte(this.PayloadIndex)
	w.WriteBool(this.Resolution)
	w.WriteByte(byte(len(this.Signatures)))
	for i := range this.MultisigPublicKeys {
		w.Write(this.MultisigPublicKeys[i])
		w.Write(this.Signatures[i])
	}
}

func (this *TransactionSimpleNothing) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if this.TxId, err = r.ReadBytes(cryptography.HashSize); err != nil {
		return
	}
	if this.PayloadIndex, err = r.ReadByte(); err != nil {
		return
	}
	if this.Resolution, err = r.ReadBool(); err != nil {
		return
	}

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}
	this.MultisigPublicKeys = make([][]byte, n)
	this.Signatures = make([][]byte, n)
	for i := range this.MultisigPublicKeys {
		if this.MultisigPublicKeys[i], err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if this.Signatures[i], err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
	}
	return
}
