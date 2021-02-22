package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	base58 "pandora-pay/helpers/base58"
)

type AddressVersion uint64

const (
	AddressVersionTransparentPublicKeyHash AddressVersion = 0
	AddressVersionTransparentPublicKey     AddressVersion = 1
)

type Address struct {
	Network   uint64
	Version   AddressVersion
	PublicKey []byte // publicKey or PublicKeyHash
	Amount    uint64 // amount to be paid
	PaymentID []byte // payment id
}

func (a *Address) EncodeAddr() (string, error) {

	var prefix string
	if a.Network == blockchain.MAIN_NET_NETWORK_BYTE {
		prefix = blockchain.MAIN_NET_NETWORK_BYTE_PREFIX
	} else if a.Network == blockchain.TEST_NET_NETWORK_BYTE {
		prefix = blockchain.TEST_NET_NETWORK_BYTE_PREFIX
	} else if a.Network == blockchain.DEV_NET_NETWORK_BYTE {
		prefix = blockchain.DEV_NET_NETWORK_BYTE_PREFIX
	} else {
		return "", errors.New("Invalid network")
	}

	bytes := append([]byte{}, helpers.SerializeNumber(uint64(a.Version))...)
	bytes = append(bytes, a.PublicKey...)

	integrationByte := a.IntegrationByte()
	bytes = append(bytes, integrationByte)

	if a.IsIntegratedAddress() {
		bytes = append(bytes, a.PaymentID...)
	}
	if a.IsIntegratedAmount() {
		bytes = append(bytes, helpers.SerializeNumber(a.Amount)...)
	}

	checksum := crypto.RIPEMD(bytes)[0:4]

	bytes = append(bytes, checksum...)

	ret := base58.Encode(bytes)

	return prefix + ret, nil
}

func DecodeAddr(input string) (*Address, error) {

	checksum := crypto.RIPEMD([]byte(input[:len(input)-4]))[0:4]

	if bytes.Equal(checksum[:], []byte(input[len(input)-4:])) {
		return nil, errors.New("Invalid Checksum")
	}
	input = input[0 : len(input)-4] // remove the checksum

	prefix := input[0:blockchain.NETWORK_BYTE_PREFIX_LENGTH]

	var network uint64
	if prefix == blockchain.MAIN_NET_NETWORK_BYTE_PREFIX {
		network = blockchain.MAIN_NET_NETWORK_BYTE
	} else if prefix == blockchain.TEST_NET_NETWORK_BYTE_PREFIX {
		network = blockchain.TEST_NET_NETWORK_BYTE
	} else if prefix == blockchain.DEV_NET_NETWORK_BYTE_PREFIX {
		network = blockchain.DEV_NET_NETWORK_BYTE
	} else {
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	raw, err := base58.Decode(input[:blockchain.NETWORK_BYTE_PREFIX_LENGTH])
	if err != nil {
		return nil, err
	}

	position := 0

	version, err := helpers.UnserializeNumber(raw, &position)
	if err != nil {
		return nil, err
	}

	var readBytes int
	addressVersion := AddressVersion(version)
	if addressVersion == AddressVersionTransparentPublicKeyHash {
		readBytes = 20
	} else if addressVersion == AddressVersionTransparentPublicKey {
		readBytes = 33
	} else {
		return nil, errors.New("Invalid Address Version")
	}

	publicKey, err := helpers.UnserializeBuffer(raw, &position, readBytes)
	if err != nil {
		return nil, err
	}

	integrationByte, err := helpers.UnserializeBuffer(raw, &position, 1)
	if err != nil {
		return nil, err
	}

	var paymentId []byte
	var amount uint64

	if integrationByte[0]&1 == 1 {

		paymentId, err = helpers.UnserializeBuffer(raw, &position, 8)
		if err != nil {
			return nil, err
		}

	}

	if integrationByte[0]&(1<<1) == 1 {

		amount, err = helpers.UnserializeNumber(raw, &position)
		if err != nil {
			return nil, err
		}

	}

	return &Address{Network: network, Version: addressVersion, PublicKey: publicKey, Amount: amount, PaymentID: paymentId}, nil
}

func (a *Address) IntegrationByte() (out byte) {

	out = 0
	if len(a.PaymentID) > 0 {
		out |= 1
	}

	if a.Amount > 0 {
		out |= 1 << 1
	}

	return
}

// tells whether address contains a paymentId
func (a *Address) IsIntegratedAddress() bool {
	return len(a.PaymentID) > 0
}

// tells whether address contains a paymentId
func (a *Address) IsIntegratedAmount() bool {
	return a.Amount > 0
}

// if address has amount
func (a Address) IntegratedAmount() uint64 {
	return a.Amount
}
