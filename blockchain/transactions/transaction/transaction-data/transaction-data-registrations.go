package transaction_data

import (
	"errors"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/helpers"
)

type TransactionDataTransactions struct {
	Registrations []*TransactionDataRegistration
}

func (self *TransactionDataTransactions) RegisterNow(regs *registrations.Registrations, publicKeyList [][]byte) (err error) {

	for _, reg := range self.Registrations {

		//verify that the other accounts did not register meanwhile
		var isReg bool
		if isReg, err = regs.Exists(string(publicKeyList[reg.PublicKeyIndex])); err != nil {
			return
		}
		if isReg {
			return errors.New("PublicKey is already registered")
		}

		//let's register
		if _, err = regs.CreateRegistration(publicKeyList[reg.PublicKeyIndex], reg.RegistrationSignature); err != nil {
			return
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
