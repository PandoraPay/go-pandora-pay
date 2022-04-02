package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/helpers/custom_base64"
)

type Address struct {
	Network        uint64         `json:"network" msgpack:"network"`
	Version        AddressVersion `json:"version" msgpack:"version"`
	PublicKey      []byte         `json:"publicKey" msgpack:"publicKey"`
	Staked         bool           `json:"staked" msgpack:"staked"`
	SpendPublicKey []byte         `json:"spendPublicKey" msgpack:"spendPublicKey"`
	Registration   []byte         `json:"registration" msgpack:"registration"`
	PaymentID      []byte         `json:"paymentId" msgpack:"paymentId"`         // payment id
	PaymentAmount  uint64         `json:"paymentAmount" msgpack:"paymentAmount"` // amount to be paid
	PaymentAsset   []byte         `json:"paymentAsset" msgpack:"paymentAsset"`
}

func NewAddr(network uint64, version AddressVersion, publicKey []byte, staked bool, spendPublicKey []byte, registration []byte, paymentID []byte, paymentAmount uint64, paymentAsset []byte) (*Address, error) {
	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Invalid PublicKey size")
	}
	if len(spendPublicKey) != 0 && len(spendPublicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Invalid Spend Public Key size")
	}
	if len(registration) != 0 && len(registration) != cryptography.SignatureSize {
		return nil, errors.New("Invalid Registration size")
	}
	if len(paymentID) != 8 && len(paymentID) != 0 {
		return nil, errors.New("Invalid PaymentID. It must be an 8 byte")
	}
	if len(paymentAsset) != 0 && len(paymentAsset) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Invalid PaymentAsset size")
	}
	return &Address{network, version, publicKey, staked, spendPublicKey, registration, paymentID, paymentAmount, paymentAsset}, nil
}

func CreateAddr(publicKey []byte, staked bool, spendPublicKey, registration []byte, paymentID []byte, paymentAmount uint64, paymentAsset []byte) (*Address, error) {
	return NewAddr(config.NETWORK_SELECTED, SIMPLE_PUBLIC_KEY, publicKey, staked, spendPublicKey, registration, paymentID, paymentAmount, paymentAsset)
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

	writer.Write(a.PublicKey)

	writer.WriteUvarint(a.IntegrationBytes())

	if a.IsIntegratedSpendPublicKey() {
		writer.Write(a.SpendPublicKey)
	}

	if a.IsIntegratedRegistration() {
		writer.Write(a.Registration)
	}
	if a.IsIntegratedPaymentID() {
		writer.Write(a.PaymentID)
	}
	if a.IsIntegratedAmount() {
		writer.WriteUvarint(a.PaymentAmount)
	}
	if a.IsIntegratedPaymentAsset() {
		writer.Write(a.PaymentAsset)
	}

	buffer := writer.Bytes()

	checksum := cryptography.GetChecksum(buffer)
	buffer = append(buffer, checksum...)
	ret := custom_base64.Base64Encoder.EncodeToString(buffer)

	return prefix + ret
}
func DecodeAddr(input string) (*Address, error) {

	addr := &Address{PublicKey: []byte{}, PaymentID: []byte{}}

	if len(input) < config.NETWORK_BYTE_PREFIX_LENGTH {
		return nil, errors.New("Invalid Address length")
	}

	prefix := input[0:config.NETWORK_BYTE_PREFIX_LENGTH]

	switch prefix {
	case config.MAIN_NET_NETWORK_BYTE_PREFIX:
		addr.Network = config.MAIN_NET_NETWORK_BYTE
	case config.TEST_NET_NETWORK_BYTE_PREFIX:
		addr.Network = config.TEST_NET_NETWORK_BYTE
	case config.DEV_NET_NETWORK_BYTE_PREFIX:
		addr.Network = config.DEV_NET_NETWORK_BYTE
	default:
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	if addr.Network != config.NETWORK_SELECTED {
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
	addr.Version = AddressVersion(version)

	switch addr.Version {
	case SIMPLE_PUBLIC_KEY:
		if addr.PublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid Address Version")
	}

	var integrationBytes uint64
	if integrationBytes, err = reader.ReadUvarint(); err != nil {
		return nil, err
	}

	addr.Staked = integrationBytes&1 != 0

	if integrationBytes&(1<<1) != 0 {
		if addr.SpendPublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
			return nil, err
		}
	}
	if integrationBytes&(1<<2) != 0 {
		if addr.Registration, err = reader.ReadBytes(cryptography.SignatureSize); err != nil {
			return nil, err
		}
	}
	if integrationBytes&(1<<3) != 0 {
		if addr.PaymentID, err = reader.ReadBytes(8); err != nil {
			return nil, err
		}
	}
	if integrationBytes&(1<<4) != 0 {
		if addr.PaymentAmount, err = reader.ReadUvarint(); err != nil {
			return nil, err
		}
	}
	if integrationBytes&(1<<5) != 0 {
		if addr.PaymentAsset, err = reader.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
			return nil, err
		}
	}

	return addr, nil
}

func (a *Address) IntegrationBytes() (out uint64) {

	out = 0

	if a.Staked {
		out |= 1
	}

	if len(a.SpendPublicKey) > 0 {
		out |= 1 << 1
	}

	if len(a.Registration) > 0 {
		out |= 1 << 2
	}

	if len(a.PaymentID) > 0 {
		out |= 1 << 3
	}

	if a.PaymentAmount > 0 {
		out |= 1 << 4
	}

	if len(a.PaymentAsset) > 0 {
		out |= 1 << 5
	}

	return
}

// if address contains a paymentId
func (a *Address) IsIntegratedRegistration() bool {
	return len(a.Registration) > 0
}

func (a *Address) IsIntegratedSpendPublicKey() bool {
	return len(a.SpendPublicKey) > 0
}

// if address contains amount
func (a *Address) IsIntegratedAmount() bool {
	return a.PaymentAmount > 0
}

// if address contains a paymentId
func (a *Address) IsIntegratedPaymentID() bool {
	return len(a.PaymentID) > 0
}

// if address contains a PaymentAsset
func (a *Address) IsIntegratedPaymentAsset() bool {
	return len(a.PaymentAsset) > 0
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
