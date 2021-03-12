package addresses

import (
	"bytes"
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
	Network   uint64
	Version   AddressVersion
	PublicKey []byte // publicKey or PublicKeyHash
	Amount    uint64 // amount to be paid
	PaymentID []byte // payment id
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

	writer.Write(a.PublicKey)

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

func DecodeAddrSilent(input string) (adr *Address, err error) {
	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()
	adr = DecodeAddr(input)
	return
}

func DecodeAddr(input string) (adr *Address) {

	adr = &Address{PublicKey: []byte{}, PaymentID: []byte{}}

	prefix := input[0:config.NETWORK_BYTE_PREFIX_LENGTH]

	switch prefix {
	case config.MAIN_NET_NETWORK_BYTE_PREFIX:
		adr.Network = config.MAIN_NET_NETWORK_BYTE
	case config.TEST_NET_NETWORK_BYTE_PREFIX:
		adr.Network = config.TEST_NET_NETWORK_BYTE
	case config.DEV_NET_NETWORK_BYTE_PREFIX:
		adr.Network = config.DEV_NET_NETWORK_BYTE
	default:
		panic("Invalid Address Network PREFIX!")
	}

	if adr.Network != config.NETWORK_SELECTED {
		panic("Address network is invalid")
	}

	var buf []byte
	var err error
	if buf, err = base58.Decode(input[config.NETWORK_BYTE_PREFIX_LENGTH:]); err != nil {
		panic("Error decoding base58")
	}

	checksum := cryptography.GetChecksum(buf[:len(buf)-cryptography.ChecksumSize])

	if !bytes.Equal(checksum, buf[len(buf)-cryptography.ChecksumSize:]) {
		panic("Invalid Checksum")
	}
	buf = buf[0 : len(buf)-cryptography.ChecksumSize] // remove the checksum

	reader := helpers.NewBufferReader(buf)

	adr.Version = AddressVersion(reader.ReadUvarint())

	var readBytes int
	switch adr.Version {
	case SimplePublicKeyHash:
		readBytes = 20
	case SimplePublicKey:
		readBytes = 33
	default:
		panic("Invalid Address Version")
	}

	adr.PublicKey = reader.ReadBytes(readBytes)
	integrationByte := reader.ReadByte()

	if integrationByte&1 != 0 {
		adr.PaymentID = reader.ReadBytes(8)
	}
	if integrationByte&(1<<1) != 0 {
		adr.Amount = reader.ReadUvarint()
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
