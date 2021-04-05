package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	base58 "pandora-pay/helpers/base58"
)

type AddressVersion uint64

const (
	SimplePublicKeyHash AddressVersion = 0
	SimplePublicKey     AddressVersion = 1
)

type Address struct {
	Network       uint64
	Version       AddressVersion
	PublicKey     helpers.HexBytes
	PublicKeyHash helpers.HexBytes
	Amount        uint64           // amount to be paid
	PaymentID     helpers.HexBytes // payment id
}

func (e AddressVersion) String() string {
	switch e {
	case SimplePublicKeyHash:
		return "Simple PubKeyHash"
	case SimplePublicKey:
		return "Simple PubKey"
	default:
		return "Unknown Address Version"
	}
}

func (a *Address) EncodeAddr() string {

	writer := helpers.NewBufferWriter()

	var prefix string
	switch a.Network {
	case config.MAIN_NET_NETWORK_BYTE:
		prefix = config.MAIN_NET_NETWORK_BYTE_PREFIX
	case config.TEST_NET_NETWORK_BYTE:
		prefix = config.TEST_NET_NETWORK_BYTE_PREFIX
	case config.DEV_NET_NETWORK_BYTE:
		prefix = config.DEV_NET_NETWORK_BYTE_PREFIX
	default:
		panic("Invalid network")
	}

	writer.WriteUvarint(uint64(a.Version))

	switch a.Version {
	case SimplePublicKey:
		writer.Write(a.PublicKey)
	case SimplePublicKeyHash:
		writer.Write(a.PublicKeyHash)
	}

	writer.WriteByte(a.IntegrationByte())

	if a.IsIntegratedAddress() {
		writer.Write(a.PaymentID)
	}
	if a.IsIntegratedAmount() {
		writer.WriteUvarint(a.Amount)
	}

	buffer := writer.Bytes()

	checksum := cryptography.GetChecksum(buffer)
	buffer = append(buffer, checksum...)
	ret := base58.Encode(buffer)

	return prefix + ret
}
func DecodeAddr(input string) (adr *Address, err error) {

	adr = &Address{PublicKey: []byte{}, PaymentID: []byte{}}

	if len(input) < config.NETWORK_BYTE_PREFIX_LENGTH {
		return nil, errors.New("Invalid Address length")
	}

	prefix := input[0:config.NETWORK_BYTE_PREFIX_LENGTH]

	switch prefix {
	case config.MAIN_NET_NETWORK_BYTE_PREFIX:
		adr.Network = config.MAIN_NET_NETWORK_BYTE
	case config.TEST_NET_NETWORK_BYTE_PREFIX:
		adr.Network = config.TEST_NET_NETWORK_BYTE
	case config.DEV_NET_NETWORK_BYTE_PREFIX:
		adr.Network = config.DEV_NET_NETWORK_BYTE
	default:
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	if adr.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Address network is invalid")
	}

	var buf []byte
	if buf, err = base58.Decode(input[config.NETWORK_BYTE_PREFIX_LENGTH:]); err != nil {
		return
	}

	checksum := cryptography.GetChecksum(buf[:len(buf)-cryptography.ChecksumSize])

	if !bytes.Equal(checksum, buf[len(buf)-cryptography.ChecksumSize:]) {
		return nil, errors.New("Invalid Checksum")
	}

	buf = buf[0 : len(buf)-cryptography.ChecksumSize] // remove the checksum

	reader := helpers.NewBufferReader(buf)

	var version uint64
	if version, err = reader.ReadUvarint(); err != nil {
		return
	}
	adr.Version = AddressVersion(version)

	switch adr.Version {
	case SimplePublicKeyHash:
		if adr.PublicKeyHash, err = reader.ReadBytes(20); err != nil {
			return
		}
	case SimplePublicKey:
		if adr.PublicKey, err = reader.ReadBytes(33); err != nil {
			return
		}
		adr.PublicKeyHash = cryptography.ComputePublicKeyHash(adr.PublicKey)
	default:
		return nil, errors.New("Invalid Address Version")
	}

	var integrationByte byte
	if integrationByte, err = reader.ReadByte(); err != nil {
		return
	}

	if integrationByte&1 != 0 {
		if adr.PaymentID, err = reader.ReadBytes(8); err != nil {
			return
		}
	}
	if integrationByte&(1<<1) != 0 {
		if adr.Amount, err = reader.ReadUvarint(); err != nil {
			return
		}
	}

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
