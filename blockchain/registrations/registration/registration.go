package registration

import (
	"pandora-pay/helpers"
)

type Registration struct {
	helpers.SerializableInterface `json:"-"`
	PublicKey                     []byte `json:"-"` //hashmap key
	Registered                    bool
}

func (registration *Registration) Serialize(w *helpers.BufferWriter) {
	w.WriteBool(registration.Registered)
}

func (registration *Registration) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	registration.Serialize(w)
	return w.Bytes()
}

func (registration *Registration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.Registered, err = r.ReadBool(); err != nil {
		return
	}
	return
}

func NewRegistration(publicKey []byte) *Registration {

	return &Registration{
		PublicKey: publicKey,
	}

	return nil
}
