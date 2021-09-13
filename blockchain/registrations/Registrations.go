package registrations

import (
	"errors"
	"pandora-pay/blockchain/registrations/registration"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	hash_map "pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Registrations struct {
	hash_map.HashMap `json:"-"`
}

func (registrations *Registrations) CreateRegistration(publicKey, registrationSignature []byte) (*registration.Registration, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	if crypto.VerifySignature([]byte("registration"), registrationSignature, publicKey) == false {
		return nil, errors.New("Registration is invalid")
	}

	reg := registration.NewRegistration(publicKey)
	if err := registrations.Update(string(publicKey), reg); err != nil {
		return nil, err
	}
	return reg, nil
}

func (registrations *Registrations) GetRegistration(key []byte) (*registration.Registration, error) {

	data, err := registrations.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	reg := data.(*registration.Registration)
	return reg, nil
}

func NewRegistrations(tx store_db_interface.StoreDBTransactionInterface) (registrations *Registrations, err error) {

	hashmap, err := hash_map.CreateNewHashMap(tx, "registrations", cryptography.PublicKeySize, true)
	if err != nil {
		return nil, err
	}

	registrations = &Registrations{
		HashMap: *hashmap,
	}

	registrations.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var reg = registration.NewRegistration(key)
		err := reg.Deserialize(helpers.NewBufferReader(data))
		return reg, err
	}
	return
}