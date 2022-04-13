package addresses

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type KeyWIF struct {
	Version  PrivateKeyVersion `json:"version,omitempty" msgpack:"version,omitempty"`
	Network  uint64            `json:"network,omitempty" msgpack:"network,omitempty"` //replay protection from one network to another one
	Key      []byte            `json:"key" msgpack:"key"`                             //32 byte
	Checksum []byte            `json:"checksum,omitempty" msgpack:"checksum,omitempty"`
}

func (pk *KeyWIF) deserialize(buffer []byte, keySize int) (err error) {

	if len(buffer) == keySize { //no wif supplied
		pk.Version = SIMPLE_PRIVATE_KEY
		pk.Network = config.MAIN_NET_NETWORK_BYTE
		pk.Key = buffer
		pk.Checksum = pk.computeCheckSum()
	} else if len(buffer) >= 1+1+keySize+cryptography.ChecksumSize { //wif

		//let's check the checksum
		checksum := cryptography.GetChecksum(buffer[:len(buffer)-cryptography.ChecksumSize])

		r := helpers.NewBufferReader(buffer)

		var version uint64
		if version, err = r.ReadUvarint(); err != nil {
			return
		}
		pk.Version = PrivateKeyVersion(version)

		if pk.Network, err = r.ReadUvarint(); err != nil {
			return
		}
		if pk.Key, err = r.ReadBytes(cryptography.PrivateKeySize); err != nil {
			return
		}
		if pk.Checksum, err = r.ReadBytes(cryptography.PrivateKeySize); err != nil {
			return
		}

		if !bytes.Equal(pk.Checksum, checksum) {
			return errors.New("Private Key WIF Checksum is not matching")
		}
	}

	return errors.New("Private Key length is invalid")
}

func (pk *KeyWIF) Serialize() []byte {
	w := helpers.NewBufferWriter()

	if pk.Version == SIMPLE_PRIVATE_KEY {
		w.Write(pk.Key)
	} else if pk.Version == SIMPLE_PRIVATE_KEY_WIF {
		w.WriteUvarint(uint64(pk.Version))
		w.WriteUvarint(pk.Network)
		w.Write(pk.Key)
		w.Write(pk.Checksum)
	}

	return w.Bytes()
}

func (pk *KeyWIF) computeCheckSum() []byte {

	w := helpers.NewBufferWriter()
	w.WriteUvarint(uint64(pk.Version))
	w.WriteUvarint(pk.Network)
	w.Write(pk.Key)

	b := w.Bytes()
	return cryptography.GetChecksum(b)
}
