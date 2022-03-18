package transaction_zether_registration

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherDataRegistration struct {
	RegistrationType           TransactionZetherDataRegistrationType
	RegistrationStaked         bool
	RegistrationSpendPublicKey []byte
	RegistrationSignature      []byte
}

func (registration *TransactionZetherDataRegistration) Serialize(w *helpers.BufferWriter) {
	w.WriteByte(byte(registration.RegistrationType))
	if registration.RegistrationType == NOT_REGISTERED {
		w.WriteBool(registration.RegistrationStaked)
		w.WriteBool(len(registration.RegistrationSpendPublicKey) > 0)
		w.Write(registration.RegistrationSpendPublicKey)
		w.Write(registration.RegistrationSignature)
	}
}

func (registration *TransactionZetherDataRegistration) Deserialize(r *helpers.BufferReader) (err error) {

	var n byte
	if n, err = r.ReadByte(); err != nil {
		return
	}

	registration.RegistrationType = TransactionZetherDataRegistrationType(n)

	switch registration.RegistrationType {
	case NOT_REGISTERED:
		if registration.RegistrationStaked, err = r.ReadBool(); err != nil {
			return
		}
		var hasSpendPublicKey bool
		if hasSpendPublicKey, err = r.ReadBool(); err != nil {
			return
		}
		if hasSpendPublicKey {
			if registration.RegistrationSpendPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
				return
			}
		}
		if registration.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
	case REGISTERED_EMPTY_ACCOUNT, REGISTERED_ACCOUNT:
		return errors.New("Registered accounts should not be manually specified")
	default:
		return errors.New("Invalid RegistrationType")
	}

	return
}
