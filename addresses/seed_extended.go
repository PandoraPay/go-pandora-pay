package addresses

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
)

type SeedExtended struct {
	KeyWIF
}

func (pke *SeedExtended) Deserialize(buffer []byte) error {
	return pke.deserialize(buffer, cryptography.SeedSize)
}

func NewSeedExtended(key []byte) (*SeedExtended, error) {

	if len(key) != cryptography.SeedSize {
		return nil, errors.New("Private Key length is invalid")
	}

	pke := &SeedExtended{
		KeyWIF{
			SIMPLE_PRIVATE_KEY_WIF,
			config.NETWORK_SELECTED,
			key,
			nil,
		},
	}

	pke.Checksum = pke.computeCheckSum()
	return pke, nil
}
