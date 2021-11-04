package registrations

import (
	"errors"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	hash_map "pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Registrations struct {
	hash_map.HashMap `json:"-"`
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

	reg := registration.NewRegistration(publicKey)
	registrations.Update(string(publicKey), reg)
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

func (registrations *Registrations) GetRandomRegistration() (*registration.Registration, error) {
	data, err := registrations.GetRandom()
	if err != nil {
		return nil, err
	}
	return data.(*registration.Registration), nil
}

func NewRegistrations(tx store_db_interface.StoreDBTransactionInterface) (registrations *Registrations) {

	hashmap := hash_map.CreateNewHashMap(tx, "registrations", cryptography.PublicKeySize, true)

	registrations = &Registrations{
		HashMap: *hashmap,
	}

	registrations.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var reg = registration.NewRegistration(key)
		err := reg.Deserialize(helpers.NewBufferReader(data))
		return reg, err
	}

	registrations.HashMap.StoredEvent = func(key []byte, element *hash_map.CommittedMapElement) error {
		element.Element.(*registration.Registration).Index = registrations.HashMap.Count
		return nil
	}

	return
}
