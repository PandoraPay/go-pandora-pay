package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/custom_base64"
)

type Address struct {
	Network       uint64         `json:"network" msgpack:"network"`
	Version       AddressVersion `json:"version" msgpack:"version"`
	PublicKeyHash []byte         `json:"publicKeyHash" msgpack:"publicKeyHash"`
	PaymentID     []byte         `json:"paymentID" msgpack:"paymentID"`         // payment id
	PaymentAmount uint64         `json:"paymentAmount" msgpack:"paymentAmount"` // amount to be paid
	PaymentAsset  []byte         `json:"paymentAsset" msgpack:"paymentAsset"`
}

func newAddr(network uint64, version AddressVersion, publicKeyHash []byte, paymentID []byte, paymentAmount uint64, paymentAsset []byte) (*Address, error) {

	if len(publicKeyHash) == cryptography.PublicKeySize {
		publicKeyHash = cryptography.GetPublicKeyHash(publicKeyHash)
	}

	if len(publicKeyHash) != cryptography.PublicKeyHashSize {
		return nil, errors.New("Invalid PublicKeyHash size")
	}
	if len(paymentID) != 8 && len(paymentID) != 0 {
		return nil, errors.New("Invalid PaymentID. It must be an 8 byte")
	}
	if len(paymentAsset) != 0 && len(paymentAsset) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Invalid PaymentAsset size")
	}
	return &Address{network, version, publicKeyHash, paymentID, paymentAmount, paymentAsset}, nil
}

func CreateAddr(publicKeyHash []byte, paymentID []byte, paymentAmount uint64, paymentAsset []byte) (*Address, error) {
	version := SIMPLE_PUBLIC_KEY_HASH
	if paymentAmount > 0 || len(paymentID) > 0 || len(paymentAsset) > 0 {
		version = SIMPLE_PUBLIC_KEY_HASH_INTEGRATED
	}
	return newAddr(config.NETWORK_SELECTED, version, publicKeyHash, paymentID, paymentAmount, paymentAsset)
}

func (a *Address) EncodeAddr() string {
	if a == nil {
		return ""
	}
	writer := helpers.NewBufferWriter()

	var prefix []byte
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

	writer.WriteByte(byte(a.Version))

	writer.Write(a.PublicKeyHash)

	if a.Version == SIMPLE_PUBLIC_KEY_HASH_INTEGRATED {
		writer.WriteUvarint(a.IntegrationBytes())

		if a.IsIntegratedPaymentID() {
			writer.Write(a.PaymentID)
		}
		if a.IsIntegratedAmount() {
			writer.WriteUvarint(a.PaymentAmount)
		}
		if a.IsIntegratedPaymentAsset() {
			writer.Write(a.PaymentAsset)
		}
	}

	buffer := writer.Bytes()

	checksum := cryptography.GetChecksum(buffer)

	final := prefix
	final = append(final, buffer...)
	final = append(final, checksum...)
	final = append(final, byte(255))

	return custom_base64.Base64Encoder.EncodeToString(final)
}
func DecodeAddr(input string) (*Address, error) {

	addr := &Address{}

	buf, err := custom_base64.Base64Encoder.DecodeString(input)
	if err != nil {
		return nil, err
	}

	if len(buf) <= config.NETWORK_BYTE_PREFIX_LENGTH {
		return nil, errors.New("Invalid Address Size")
	}

	prefix := buf[:config.NETWORK_BYTE_PREFIX_LENGTH]
	if bytes.Equal(prefix, config.MAIN_NET_NETWORK_BYTE_PREFIX) {
		addr.Network = config.MAIN_NET_NETWORK_BYTE
	} else if bytes.Equal(prefix, config.TEST_NET_NETWORK_BYTE_PREFIX) {
		addr.Network = config.TEST_NET_NETWORK_BYTE
	} else if bytes.Equal(prefix, config.DEV_NET_NETWORK_BYTE_PREFIX) {
		addr.Network = config.DEV_NET_NETWORK_BYTE
	} else {
		return nil, errors.New("Invalid Address Network PREFIX!")
	}

	if addr.Network != config.NETWORK_SELECTED {
		return nil, errors.New("Address network is invalid")
	}

	if buf[len(buf)-1] != 255 {
		return nil, errors.New("Suffix is not matching")
	}

	checksum := cryptography.GetChecksum(buf[config.NETWORK_BYTE_PREFIX_LENGTH : len(buf)-cryptography.ChecksumSize-1])

	if !bytes.Equal(checksum, buf[len(buf)-cryptography.ChecksumSize-1:len(buf)-1]) {
		return nil, errors.New("Invalid Checksum")
	}

	buf = buf[config.NETWORK_BYTE_PREFIX_LENGTH : len(buf)-cryptography.ChecksumSize-1] // remove the checksum

	reader := helpers.NewBufferReader(buf)

	var version byte
	if version, err = reader.ReadByte(); err != nil {
		return nil, err
	}
	addr.Version = AddressVersion(version)

	switch addr.Version {
	case SIMPLE_PUBLIC_KEY_HASH, SIMPLE_PUBLIC_KEY_HASH_INTEGRATED:
		if addr.PublicKeyHash, err = reader.ReadBytes(cryptography.PublicKeyHashSize); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid Address Version")
	}

	if addr.Version == SIMPLE_PUBLIC_KEY_HASH_INTEGRATED {
		var integrationBytes uint64
		if integrationBytes, err = reader.ReadUvarint(); err != nil {
			return nil, err
		}

		if integrationBytes&1 != 0 {
			if addr.PaymentID, err = reader.ReadBytes(8); err != nil {
				return nil, err
			}
		}
		if integrationBytes&(1<<1) != 0 {
			if addr.PaymentAmount, err = reader.ReadUvarint(); err != nil {
				return nil, err
			}
		}
		if integrationBytes&(1<<2) != 0 {
			if addr.PaymentAsset, err = reader.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
				return nil, err
			}
		}
	}

	return addr, nil
}

func (a *Address) IntegrationBytes() (out uint64) {

	out = 0

	if len(a.PaymentID) > 0 {
		out |= 1
	}

	if a.PaymentAmount > 0 {
		out |= 1 << 1
	}

	if len(a.PaymentAsset) > 0 {
		out |= 1 << 2
	}

	return
}

// if address contains amount
func (a *Address) IsIntegratedAmount() bool {
	return a.PaymentAmount > 0
}

// if address contains a paymentID
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
