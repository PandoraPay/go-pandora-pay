package transaction_zether_registrations

import (
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
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
		if reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED {
			publicKey := publickeylist[i]
			if crypto.VerifySignaturePoint([]byte("registration"), reg.RegistrationSignature, publicKey) == false {
				return fmt.Errorf("Registration is invalid for %d", i)
			}
		}
	}

	return
}

func (self *TransactionZetherDataRegistrations) RegisterNow(asset []byte, dataStorage *data_storage.DataStorage, publicKeyList [][]byte) (err error) {

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(asset); err != nil {
		return
	}

	var isReg bool
	for i, reg := range self.Registrations {

		if reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED {
			//verify that the other accounts did not register meanwhile
			if isReg, err = dataStorage.Regs.Exists(string(publicKeyList[i])); err != nil {
				return
			}
			if isReg {
				return errors.New("PublicKey is already registered")
			}

			//let's register
			if _, err = dataStorage.Regs.CreateRegistration(publicKeyList[i]); err != nil {
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

		if reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED || reg.RegistrationType == transaction_zether_registration.REGISTERED_EMPTY_ACCOUNT {

			var exists bool
			if exists, err = accs.Exists(string(publicKeyList[i])); err != nil {
				return
			}

			if exists {
				return errors.New("Account is already registered")
			}

			if _, err = accs.CreateAccount(publicKeyList[i]); err != nil {
				return
			}

		}
	}

	return
}

func (self *TransactionZetherDataRegistrations) Serialize(w *helpers.BufferWriter) {
	for _, registration := range self.Registrations {
		registration.Serialize(w)
	}
}

func (self *TransactionZetherDataRegistrations) Deserialize(r *helpers.BufferReader, ringSize int) (err error) {

	self.Registrations = make([]*transaction_zether_registration.TransactionZetherDataRegistration, ringSize)
	for i := 0; i < ringSize; i++ {
		self.Registrations[i] = &transaction_zether_registration.TransactionZetherDataRegistration{}
		if err = self.Registrations[i].Deserialize(r); err != nil {
			return
		}
	}
	return
}
