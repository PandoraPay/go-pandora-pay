package addresses

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
)

type PrivateKeyExtended struct {
	KeyWIF
}

func (pke *PrivateKeyExtended) Deserialize(buffer []byte) error {
	return pke.deserialize(buffer, cryptography.PrivateKeySizeExtended)
}

func NewPrivateKeyExtended(key []byte) (*PrivateKeyExtended, error) {

	if len(key) != cryptography.PrivateKeySizeExtended {
		return nil, errors.New("Private Key length is invalid")
	}

	version := SIMPLE_PRIVATE_KEY_WIF
	network := config.NETWORK_SELECTED

	privateKeyExtended := &PrivateKeyExtended{
		KeyWIF{
			version,
			network,
			key,
			nil,
		},
	}

	privateKeyExtended.Checksum = privateKeyExtended.computeCheckSum()
	return privateKeyExtended, nil
}
