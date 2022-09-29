package registrations

import (
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Registrations struct {
	*hash_map.HashMap[*registration.Registration]
}

func VerifyRegistration(publicKey []byte, staked bool, spendPublicKey, registrationSignature []byte) bool {
	data := []byte("registration")
	if staked {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	data = append(data, spendPublicKey...)
	return crypto.VerifySignature(data, registrationSignature, publicKey)
}

func VerifyRegistrationPoint(publicKey *bn256.G1, staked bool, spendPublicKey, registrationSignature []byte) bool {
	data := []byte("registration")
	if staked {
		data = append(data, 1)
	} else {
		data = append(data, 0)
	}
	data = append(data, spendPublicKey...)
	return crypto.VerifySignaturePoint(data, registrationSignature, publicKey)
}

// WARNING: should NOT be used manually without being called from DataStorage
func (this *Registrations) CreateNewRegistration(publicKey []byte, staked bool, spendPublicKey []byte) (*registration.Registration, error) {
	reg := registration.NewRegistration(publicKey, 0) //index will be set by update
	reg.Staked = staked
	reg.SpendPublicKey = spendPublicKey
	if err := this.HashMap.Create(string(publicKey), reg); err != nil {
		return nil, err
	}
	return reg, nil
}

func NewRegistrations(tx store_db_interface.StoreDBTransactionInterface) (this *Registrations) {

	this = &Registrations{
		hash_map.CreateNewHashMap[*registration.Registration](tx, "registrations", cryptography.PublicKeySize, true),
	}

	this.HashMap.CreateObject = func(key []byte, index uint64) (*registration.Registration, error) {
		return registration.NewRegistration(key, index), nil
	}

	return
}
