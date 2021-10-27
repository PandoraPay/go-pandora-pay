package transaction_zether_registrations

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherDataRegistrations struct {
	Registrations []*TransactionZetherDataRegistration
}

func (self *TransactionZetherDataRegistrations) ValidateRegistrations(publickeylist []*bn256.G1) (err error) {

	for _, reg := range self.Registrations {

		if reg.PublicKeyIndex >= byte(len(publickeylist)) {
			return fmt.Errorf("reg.PublicKeyIndex %d exceeds %d ", reg.PublicKeyIndex, len(publickeylist))
		}

		publicKey := publickeylist[reg.PublicKeyIndex]
		if crypto.VerifySignaturePoint([]byte("registration"), reg.RegistrationSignature, publicKey) == false {
			return fmt.Errorf("Registration is invalid for %d", reg.PublicKeyIndex)
		}

	}

	return
}

func (self *TransactionZetherDataRegistrations) RegisterNow(dataStorage *data_storage.DataStorage, publicKeyList [][]byte) (err error) {

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

func (self *TransactionZetherDataRegistrations) Serialize(w *helpers.BufferWriter) {
	w.WriteByte(byte(len(self.Registrations)))
	for _, registration := range self.Registrations {
		registration.Serialize(w)
	}
}

func (self *TransactionZetherDataRegistrations) Deserialize(r *helpers.BufferReader) (err error) {

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}

	self.Registrations = make([]*TransactionZetherDataRegistration, n)
	for i := byte(0); i < n; i++ {
		self.Registrations[i] = &TransactionZetherDataRegistration{}
		if err = self.Registrations[i].Deserialize(r); err != nil {
			return
		}
	}
	return
}
