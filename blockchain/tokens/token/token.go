package token

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type Token struct {
	Version uint64
	//upgrade different settings
	canUpgrade bool
	//increase supply
	canMint bool
	//decrease supply
	canBurn bool
	//can change key
	canChangeKey bool
	//can pause (suspend transactions)
	canPause bool
	//freeze supply changes
	canFreeze bool

	decimalSeparator byte
	maxSupply        uint64
	supply           uint64

	key       [20]byte
	supplyKey [20]byte

	name        string
	ticker      string
	description string
}

func (token *Token) Serialize() []byte {

	var serialized bytes.Buffer
	temp := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(temp, token.Version)
	serialized.Write(temp[:n])

	serialized.WriteByte(helpers.SerializeBoolToByte(token.canUpgrade))
	serialized.WriteByte(helpers.SerializeBoolToByte(token.canMint))
	serialized.WriteByte(helpers.SerializeBoolToByte(token.canBurn))
	serialized.WriteByte(helpers.SerializeBoolToByte(token.canChangeKey))
	serialized.WriteByte(helpers.SerializeBoolToByte(token.canPause))
	serialized.WriteByte(helpers.SerializeBoolToByte(token.canFreeze))
	serialized.WriteByte(token.decimalSeparator)

	n = binary.PutUvarint(temp, token.maxSupply)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, token.supply)
	serialized.Write(temp[:n])

	serialized.Write(token.key[:])
	serialized.Write(token.supplyKey[:])

	serialized.Write([]byte(token.name))
	serialized.Write([]byte(token.ticker))
	serialized.Write([]byte(token.description))

	return serialized.Bytes()
}

func (token *Token) Deserialize(buf []byte) (err error) {
	return
}
