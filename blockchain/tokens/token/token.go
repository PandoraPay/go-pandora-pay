package token

import (
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

	writer := helpers.NewBufferWriter()

	writer.WriteUint64(token.Version)

	writer.WriteBool(token.canUpgrade)
	writer.WriteBool(token.canMint)
	writer.WriteBool(token.canBurn)
	writer.WriteBool(token.canChangeKey)
	writer.WriteBool(token.canPause)
	writer.WriteBool(token.canFreeze)
	writer.WriteByte(token.decimalSeparator)

	writer.WriteUint64(token.maxSupply)

	writer.WriteUint64(token.supply)

	writer.Write(token.key[:])
	writer.Write(token.supplyKey[:])

	writer.Write([]byte(token.name))
	writer.Write([]byte(token.ticker))
	writer.Write([]byte(token.description))

	return writer.Bytes()
}

func (token *Token) Deserialize(buf []byte) (err error) {
	return
}
