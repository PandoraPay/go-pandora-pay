package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/helpers/custom_base64"
)

type Address struct {
	Network      uint64           `json:"network"`
	Version      AddressVersion   `json:"version"`
	PublicKey    helpers.HexBytes `json:"publicKey"`
	Registration helpers.HexBytes `json:"registration"`
	Amount       uint64           `json:"amount"`    // amount to be paid
	PaymentID    helpers.HexBytes `json:"paymentId"` // payment id
}

func NewAddr(network uint64, version AddressVersion, publicKey []byte, registration []byte, amount uint64, paymentID []byte) (*Address, error) {
	if len(paymentID) != 8 && len(paymentID) != 0 {
		return nil, errors.New("Invalid PaymentId. It must be an 8 byte")
	}
	return &Address{network, version, publicKey, registration, amount, paymentID}, nil
}

func CreateAddr(key, registration []byte, amount uint64, paymentID []byte) (*Address, error) {

	var publicKey []byte

	var version AddressVersion
	switch len(key) {
	case cryptography.PublicKeySize:
		publicKey = key
		version = SIMPLE_PUBLIC_KEY
	default:
		return nil, errors.New("Invalid Key length")
	}

	return NewAddr(config.NETWORK_SELECTED, version, publicKey, registration, amount, paymentID)
}

func (a *Address) EncodeAddr() string {
	if a == nil {
		return ""
	}

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
	case SIMPLE_PUBLIC_KEY:
		writer.Write(a.PublicKey)
	}

	writer.WriteByte(a.IntegrationByte())

	if a.IsIntegratedRegistration() {
		writer.Write(a.Registration)
	}
	if a.IsIntegratedPaymentID() {
		writer.Write(a.PaymentID)
	}
	if a.IsIntegratedAmount() {
		writer.WriteUvarint(a.Amount)
	}

	buffer := writer.Bytes()

	checksum := cryptography.GetChecksum(buffer)
	buffer = append(buffer, checksum...)
	ret := custom_base64.Base64Encoder.EncodeToString(buffer)

	return prefix + ret
}
func DecodeAddr(input string) (*Address, error) {

	adr := &Address{PublicKey: []byte{}, PaymentID: []byte{}}

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

	buf, err := custom_base64.Base64Encoder.DecodeString(input[config.NETWORK_BYTE_PREFIX_LENGTH:])
	if err != nil {
		return nil, err
	}

	checksum := cryptography.GetChecksum(buf[:len(buf)-cryptography.ChecksumSize])

	if !bytes.Equal(checksum, buf[len(buf)-cryptography.ChecksumSize:]) {
		return nil, errors.New("Invalid Checksum")
	}

	buf = buf[0 : len(buf)-cryptography.ChecksumSize] // remove the checksum

	reader := helpers.NewBufferReader(buf)

	var version uint64
	if version, err = reader.ReadUvarint(); err != nil {
		return nil, err
	}
	adr.Version = AddressVersion(version)

	switch adr.Version {
	case SIMPLE_PUBLIC_KEY:
		if adr.PublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid Address Version")
	}

	var integrationByte byte
	if integrationByte, err = reader.ReadByte(); err != nil {
		return nil, err
	}

	if integrationByte&1 != 0 {
		if adr.Registration, err = reader.ReadBytes(cryptography.SignatureSize); err != nil {
			return nil, err
		}
	}
	if integrationByte&(1<<1) != 0 {
		if adr.PaymentID, err = reader.ReadBytes(8); err != nil {
			return nil, err
		}
	}
	if integrationByte&(1<<2) != 0 {
		if adr.Amount, err = reader.ReadUvarint(); err != nil {
			return nil, err
		}
	}

	return adr, nil
}

func (a *Address) IntegrationByte() (out byte) {

	out = 0

	if len(a.Registration) > 0 {
		out |= 1
	}

	if len(a.PaymentID) > 0 {
		out |= 1 << 1
	}

	if a.Amount > 0 {
		out |= 1 << 2
	}

	return
}

// tells whether address contains a paymentId
func (a *Address) IsIntegratedRegistration() bool {
	return len(a.Registration) > 0
}

// tells whether address contains a paymentId
func (a *Address) IsIntegratedPaymentID() bool {
	return len(a.PaymentID) > 0
}

// tells whether address contains amount
func (a *Address) IsIntegratedAmount() bool {
	return a.Amount > 0
}

func (a *Address) EncryptMessage(message []byte) ([]byte, error) {
	panic("not implemented")
}

func (a *Address) VerifySignedMessage(message, signature []byte) bool {
	return crypto.VerifySignature(message, signature, a.PublicKey)
}

func (a *Address) GetPoint() (*crypto.Point, error) {
	var point crypto.Point
	var err error

	if err = point.DecodeCompressed(a.PublicKey); err != nil {
		return nil, err
	}

	return &point, nil
}
