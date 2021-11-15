package registrations

import (
	"errors"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Registrations struct {
	*hash_map.HashMap
}

func (registrations *Registrations) VerifyRegistration(publicKey, registrationSignature []byte) bool {
	return crypto.VerifySignature([]byte("registration"), registrationSignature, publicKey)
}

func (registrations *Registrations) VerifyRegistrationPoint(publicKey *bn256.G1, registrationSignature []byte) bool {
	return crypto.VerifySignaturePoint([]byte("registration"), registrationSignature, publicKey)
}

func (registrations *Registrations) CreateRegistration(publicKey []byte) (*registration.Registration, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	reg := registration.NewRegistration(publicKey, 0) //index will be set by update
	if err := registrations.HashMap.Update(string(publicKey), reg); err != nil {
		return nil, err
	}
	return reg, nil
}

func (registrations *Registrations) GetRegistration(key []byte) (*registration.Registration, error) {

	data, err := registrations.HashMap.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	reg := data.(*registration.Registration)
	return reg, nil
}

func (registrations *Registrations) GetRandomRegistration() (*registration.Registration, error) {
	data, err := registrations.HashMap.GetRandom()
	if err != nil {
		return nil, err
	}
	return data.(*registration.Registration), nil
}

func NewRegistrations(tx store_db_interface.StoreDBTransactionInterface) (registrations *Registrations) {

	hashmap := hash_map.CreateNewHashMap(tx, "registrations", cryptography.PublicKeySize, true)

	registrations = &Registrations{
		HashMap: hashmap,
	}

	registrations.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return registration.NewRegistration(key, index), nil
	}

	registrations.HashMap.StoredEvent = func(key []byte, element *hash_map.CommittedMapElement) error {
		return nil
	}

	return
}
