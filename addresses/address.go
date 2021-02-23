package addresses

import (
	"bytes"
	"encoding/binary"
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

	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

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

	n := binary.PutUvarint(buf, uint64(a.Version))
	serialized.Write(buf[:n])

	serialized.Write(a.PublicKey)

	integrationByte := a.IntegrationByte()
	serialized.Write([]byte{integrationByte})

	if a.IsIntegratedAddress() {
		serialized.Write(a.PaymentID)
	}
	if a.IsIntegratedAmount() {
		n = binary.PutUvarint(buf, a.Amount)
		serialized.Write(buf[:n])
	}

	buffer := serialized.Bytes()

	checksum := crypto.RIPEMD(buffer)[0:crypto.ChecksumSize]
	buffer = append(buffer, checksum...)
	ret := base58.Encode(buffer)

	return prefix + ret, nil
}

func DecodeAddr(input string) (*Address, error) {

	adr := Address{PublicKey: []byte{}, PaymentID: []byte{}}

	prefix := input[0:blockchain.NETWORK_BYTE_PREFIX_LENGTH]

	if prefix == blockchain.MAIN_NET_NETWORK_BYTE_PREFIX {
		adr.Network = blockchain.MAIN_NET_NETWORK_BYTE
	} else if prefix == blockchain.TEST_NET_NETWORK_BYTE_PREFIX {
		adr.Network = blockchain.TEST_NET_NETWORK_BYTE
	} else if prefix == blockchain.DEV_NET_NETWORK_BYTE_PREFIX {
		adr.Network = blockchain.DEV_NET_NETWORK_BYTE
	} else {
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	buf, err := base58.Decode(input[blockchain.NETWORK_BYTE_PREFIX_LENGTH:])
	if err != nil {
		return nil, err
	}

	checksum := crypto.RIPEMD(buf[:len(buf)-crypto.ChecksumSize])[0:crypto.ChecksumSize]

	if !bytes.Equal(checksum[:], buf[len(buf)-crypto.ChecksumSize:]) {
		return nil, errors.New("Invalid Checksum")
	}
	buf = buf[0 : len(buf)-crypto.ChecksumSize] // remove the checksum

	var version uint64

	version, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return nil, err
	}
	adr.Version = AddressVersion(version)

	var readBytes int

	switch adr.Version {
	case AddressVersionTransparentPublicKeyHash:
		readBytes = 20
	case AddressVersionTransparentPublicKey:
		readBytes = 33
	default:
		return nil, errors.New("Invalid Address Version")
	}

	adr.PublicKey, buf, err = helpers.DeserializeBuffer(buf, readBytes)
	if err != nil {
		return nil, err
	}

	var integrationByte []byte
	integrationByte, buf, err = helpers.DeserializeBuffer(buf, 1)
	if err != nil {
		return nil, err
	}

	if integrationByte[0]&1 != 0 {
		adr.PaymentID, buf, err = helpers.DeserializeBuffer(buf, 8)
		if err != nil {
			return nil, err
		}
	}

	if integrationByte[0]&(1<<1) != 0 {

		adr.Amount, buf, err = helpers.DeserializeNumber(buf)
		if err != nil {
			return nil, err
		}

	}

	return &adr, nil
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
