package transaction_data

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionDataTransactions struct {
	Registrations []*TransactionDataRegistration
}

func (self *TransactionDataTransactions) ValidateRegistrations(publickeylist []*bn256.G1) (err error) {

	for _, reg := range self.Registrations {

		if reg.PublicKeyIndex >= uint64(len(publickeylist)) {
			return fmt.Errorf("reg.PublicKeyIndex %d exceeds %d ", reg.PublicKeyIndex, len(publickeylist))
		}

		publicKey := publickeylist[reg.PublicKeyIndex]
		if crypto.VerifySignaturePoint([]byte("registration"), reg.RegistrationSignature, publicKey) == false {
			return fmt.Errorf("Registration is invalid for %d", reg.PublicKeyIndex)
		}

	}

	return
}

func (self *TransactionDataTransactions) RegisterNow(dataStorage *data_storage.DataStorage, publicKeyList [][]byte) (err error) {

	var isReg bool
	for _, reg := range self.Registrations {

		//verify that the other accounts did not register meanwhile
		if isReg, err = dataStorage.Regs.Exists(string(publicKeyList[reg.PublicKeyIndex])); err != nil {
			return
		}
		if isReg {
			return errors.New("PublicKey is already registered")
		}

		//let's register
		if _, err = dataStorage.Regs.CreateRegistration(publicKeyList[reg.PublicKeyIndex], reg.RegistrationSignature); err != nil {
			return
		}
	}

	for _, publicKey := range publicKeyList {
		if isReg, err = dataStorage.Regs.Exists(string(publicKey)); err != nil {
			return
		}
		if !isReg {
			return errors.New("PublicKey is already registered")
		}
	}

	return
}

func (self *TransactionDataTransactions) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(uint64(len(self.Registrations)))
	for _, registration := range self.Registrations {
		registration.Serialize(w)
	}
}

func (self *TransactionDataTransactions) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	self.Registrations = make([]*TransactionDataRegistration, n)
	for i := uint64(0); i < n; i++ {
		registration := &TransactionDataRegistration{}
		if err = registration.Deserialize(r); err != nil {
			return
		}
		self.Registrations[i] = registration
	}
	return
}
