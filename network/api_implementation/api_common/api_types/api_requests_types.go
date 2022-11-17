package api_types

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type APIAccountBaseRequest struct {
	Address   string         `json:"address,omitempty" msgpack:"address,omitempty"`
	PublicKey helpers.Base64 `json:"publicKey,omitempty"  msgpack:"publicKey,omitempty"`
}

func (request *APIAccountBaseRequest) GetPublicKey(required bool) ([]byte, error) {
	if request == nil {
		return nil, errors.New("argument missing")
	}

	var publicKey []byte
	if request.Address != "" {
		address, err := addresses.DecodeAddr(request.Address)
		if err != nil {
			return nil, errors.New("Invalid address")
		}
		publicKey = address.PublicKey
	} else if request.PublicKey != nil && len(request.PublicKey) == cryptography.PublicKeySize {
		publicKey = request.PublicKey
	} else if required {
		return nil, errors.New("Invalid address or publicKey")
	}

	return publicKey, nil
}
