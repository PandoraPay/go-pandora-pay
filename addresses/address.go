package addresses

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/crypto"
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

	var serialised bytes.Buffer
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
	serialised.Write(buf[:n])

	serialised.Write(a.PublicKey)

	integrationByte := a.IntegrationByte()
	serialised.Write([]byte{integrationByte})

	if a.IsIntegratedAddress() {
		serialised.Write(a.PaymentID)
	}
	if a.IsIntegratedAmount() {
		n = binary.PutUvarint(buf, a.Amount)
		serialised.Write(buf[:n])
	}

	buffer := serialised.Bytes()

	checksum := crypto.RIPEMD(buffer)[0:4]
	buffer = append(buffer, checksum...)
	ret := base58.Encode(buffer)

	return prefix + ret, nil
}

func DecodeAddr(input string) (*Address, error) {

	adr := Address{}

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

	checksum := crypto.RIPEMD(buf[:len(buf)-4])[0:4]

	if !bytes.Equal(checksum[:], buf[len(buf)-4:]) {
		return nil, errors.New("Invalid Checksum")
	}
	buf = buf[0 : len(buf)-4] // remove the checksum

	version, n := binary.Uvarint(buf)
	if n <= 0 {
		return nil, err
	}
	buf = buf[n:]
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

	adr.PublicKey = buf[0:readBytes]
	buf = buf[readBytes:]

	integrationByte := buf[0]
	buf = buf[1:]

	if integrationByte&1 != 0 {

		adr.PaymentID = buf[0:8]
		buf = buf[8:]

	}

	if integrationByte&(1<<1) != 0 {

		adr.Amount, n = binary.Uvarint(buf)
		if n <= 0 {
			return nil, errors.New("Invalid amount")
		}
		buf = buf[n:]

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
