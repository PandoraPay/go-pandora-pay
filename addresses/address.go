package addresses

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	base58 "pandora-pay/helpers/base58"
)

type AddressVersion uint64

const (
	TransparentPublicKeyHash AddressVersion = 0
	TransparentPublicKey     AddressVersion = 1
)

type Address struct {
	Network   uint64
	Version   AddressVersion
	PublicKey []byte // publicKey or PublicKeyHash
	Amount    uint64 // amount to be paid
	PaymentID []byte // payment id
}

func (e AddressVersion) String() string {
	switch e {
	case TransparentPublicKeyHash:
		return "Transparent PubKeyHash"
	case TransparentPublicKey:
		return "Transparent PubKey"
	default:
		return "Unknown Address Version"
	}
}

func (a *Address) EncodeAddr() (string, error) {

	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	var prefix string
	switch a.Network {
	case config.MAIN_NET_NETWORK_BYTE:
		prefix = config.MAIN_NET_NETWORK_BYTE_PREFIX
	case config.TEST_NET_NETWORK_BYTE:
		prefix = config.TEST_NET_NETWORK_BYTE_PREFIX
	case config.DEV_NET_NETWORK_BYTE:
		prefix = config.DEV_NET_NETWORK_BYTE_PREFIX
	default:
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

	checksum := crypto.RIPEMD(buffer)[0:helpers.ChecksumSize]
	buffer = append(buffer, checksum...)
	ret := base58.Encode(buffer)

	return prefix + ret, nil
}

func DecodeAddr(input string) (addr2 *Address, err error) {

	adr := Address{PublicKey: []byte{}, PaymentID: []byte{}}

	prefix := input[0:config.NETWORK_BYTE_PREFIX_LENGTH]

	if prefix == config.MAIN_NET_NETWORK_BYTE_PREFIX {
		adr.Network = config.MAIN_NET_NETWORK_BYTE
	} else if prefix == config.TEST_NET_NETWORK_BYTE_PREFIX {
		adr.Network = config.TEST_NET_NETWORK_BYTE
	} else if prefix == config.DEV_NET_NETWORK_BYTE_PREFIX {
		adr.Network = config.DEV_NET_NETWORK_BYTE
	} else {
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	if adr.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Address network is invalid")
	}

	var buf []byte
	if buf, err = base58.Decode(input[config.NETWORK_BYTE_PREFIX_LENGTH:]); err != nil {
		return
	}

	checksum := crypto.RIPEMD(buf[:len(buf)-helpers.ChecksumSize])[0:helpers.ChecksumSize]

	if !bytes.Equal(checksum[:], buf[len(buf)-helpers.ChecksumSize:]) {
		return nil, errors.New("Invalid Checksum")
	}
	buf = buf[0 : len(buf)-helpers.ChecksumSize] // remove the checksum

	var version uint64

	if version, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}
	adr.Version = AddressVersion(version)

	var readBytes int

	switch adr.Version {
	case TransparentPublicKeyHash:
		readBytes = 20
	case TransparentPublicKey:
		readBytes = 33
	default:
		return nil, errors.New("Invalid Address Version")
	}

	if adr.PublicKey, buf, err = helpers.DeserializeBuffer(buf, readBytes); err != nil {
		return
	}

	var integrationByte []byte
	if integrationByte, buf, err = helpers.DeserializeBuffer(buf, 1); err != nil {
		return
	}

	if integrationByte[0]&1 != 0 {
		if adr.PaymentID, buf, err = helpers.DeserializeBuffer(buf, 8); err != nil {
			return
		}
	}

	if integrationByte[0]&(1<<1) != 0 {

		if adr.Amount, buf, err = helpers.DeserializeNumber(buf); err != nil {
			return
		}

	}

	addr2 = &adr
	return
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
