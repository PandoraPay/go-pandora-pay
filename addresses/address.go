package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	base58 "pandora-pay/helpers/base58"
)

type Address struct {
	Network       uint64           `json:"network"`
	Version       AddressVersion   `json:"version"`
	PublicKey     helpers.HexBytes `json:"publicKey"`
	PublicKeyHash helpers.HexBytes `json:"publicKeyHash"`
	Amount        uint64           `json:"amount"`    // amount to be paid
	PaymentID     helpers.HexBytes `json:"paymentId"` // payment id
}

func NewAddr(network uint64, version AddressVersion, publicKey []byte, publicKeyHash []byte, amount uint64, paymentID []byte) (*Address, error) {
	if len(paymentID) != 8 && len(paymentID) != 0 {
		return nil, errors.New("Invalid PaymentId. It must be an 8 byte")
	}
	return &Address{network, version, publicKey, publicKeyHash, amount, paymentID}, nil
}

func CreateAddr(key []byte, amount uint64, paymentID []byte) (*Address, error) {

	var publicKey, publicKeyHash []byte

	var version AddressVersion
	switch len(key) {
	case cryptography.PublicKeyHashHashSize:
		publicKeyHash = key
		version = SIMPLE_PUBLIC_KEY_HASH
	case cryptography.PublicKeySize:
		publicKey = key
		version = SIMPLE_PUBLIC_KEY
	default:
		return nil, errors.New("Invalid Key length")
	}

	return NewAddr(config.NETWORK_SELECTED, version, publicKey, publicKeyHash, amount, paymentID)
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
	case SIMPLE_PUBLIC_KEY:
		writer.Write(a.PublicKey)
	case SIMPLE_PUBLIC_KEY_HASH:
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

	buf, err := base58.Decode(input[config.NETWORK_BYTE_PREFIX_LENGTH:])
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
	case SIMPLE_PUBLIC_KEY_HASH:
		if adr.PublicKeyHash, err = reader.ReadBytes(cryptography.PublicKeyHashHashSize); err != nil {
			return nil, err
		}
	case SIMPLE_PUBLIC_KEY:
		if adr.PublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
			return nil, err
		}
		adr.PublicKeyHash = cryptography.ComputePublicKeyHash(adr.PublicKey)
	default:
		return nil, errors.New("Invalid Address Version")
	}

	var integrationByte byte
	if integrationByte, err = reader.ReadByte(); err != nil {
		return nil, err
	}

	if integrationByte&1 != 0 {
		if adr.PaymentID, err = reader.ReadBytes(8); err != nil {
			return nil, err
		}
	}
	if integrationByte&(1<<1) != 0 {
		if adr.Amount, err = reader.ReadUvarint(); err != nil {
			return nil, err
		}
	}

	return adr, nil
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
