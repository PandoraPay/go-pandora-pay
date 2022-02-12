package transaction_zether_registrations

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherDataRegistrations struct {
	Registrations []*transaction_zether_registration.TransactionZetherDataRegistration
}

func (self *TransactionZetherDataRegistrations) ValidateRegistrations(publickeylist []*bn256.G1) (err error) {

	if len(publickeylist) == 0 || len(publickeylist) > config.TRANSACTIONS_ZETHER_RING_MAX {
		return errors.New("Invalid PublicKeys length")
	}

	for i, reg := range self.Registrations {
		if reg != nil && reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED {
			publicKey := publickeylist[i]
			if crypto.VerifySignaturePoint([]byte("registration"), reg.RegistrationSignature, publicKey) == false {
				return fmt.Errorf("Registration is invalid for %d", i)
			}
		}
	}

	return
}

func (self *TransactionZetherDataRegistrations) RegisterNow(asset []byte, dataStorage *data_storage.DataStorage, publicKeyList [][]byte) (err error) {

	var isReg bool
	for i, reg := range self.Registrations {
		if reg != nil && reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED {
			if _, err = dataStorage.CreateRegistration(publicKeyList[i]); err != nil {
				return
			}
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

	for i, reg := range self.Registrations {
		if reg != nil && (reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED || reg.RegistrationType == transaction_zether_registration.REGISTERED_EMPTY_ACCOUNT) {
			if _, err = dataStorage.CreateAccount(asset, publicKeyList[i]); err != nil {
				return
			}
		}
	}

	return
}

func (self *TransactionZetherDataRegistrations) Serialize(w *helpers.BufferWriter) {

	count := uint64(0)
	for _, registration := range self.Registrations {
		if registration != nil {
			count += 1
		}
	}

	w.WriteUvarint(count)
	for i, registration := range self.Registrations {
		if registration != nil {
			w.WriteUvarint(uint64(i))
			registration.Serialize(w)
		}
	}
}

func (self *TransactionZetherDataRegistrations) Deserialize(r *helpers.BufferReader, ringSize int) (err error) {

	var count uint64
	if count, err = r.ReadUvarint(); err != nil {
		return
	}

	self.Registrations = make([]*transaction_zether_registration.TransactionZetherDataRegistration, ringSize)

	for i := uint64(0); i < count; i++ {

		var index uint64
		if index, err = r.ReadUvarint(); err != nil {
			return
		}

		if index >= uint64(ringSize) {
			return errors.New("Registration Index is invalid")
		}
		if self.Registrations[index] != nil {
			return errors.New("Registration already exists")
		}

		self.Registrations[index] = &transaction_zether_registration.TransactionZetherDataRegistration{}
		if err = self.Registrations[index].Deserialize(r); err != nil {
			return
		}
	}
	return
}
