package token

import (
	"pandora-pay/helpers"
)

type Token struct {
	Version uint64
	//upgrade different settings
	CanUpgrade bool
	//increase supply
	CanMint bool
	//decrease supply
	CanBurn bool
	//can change key
	CanChangeKey bool
	//can change supply key
	CanChangeSupplyKey bool
	//can pause (suspend transactions)
	CanPause bool
	//freeze supply changes
	CanFreeze bool

	DecimalSeparator byte
	MaxSupply        uint64
	Supply           uint64

	Key       [20]byte
	SupplyKey [20]byte

	Name        string
	Ticker      string
	Description string
}

func (token *Token) Serialize() []byte {

	writer := helpers.NewBufferWriter()

	writer.WriteUint64(token.Version)

	writer.WriteBool(token.CanUpgrade)
	writer.WriteBool(token.CanMint)
	writer.WriteBool(token.CanBurn)
	writer.WriteBool(token.CanChangeKey)
	writer.WriteBool(token.CanChangeSupplyKey)
	writer.WriteBool(token.CanPause)
	writer.WriteBool(token.CanFreeze)
	writer.WriteByte(token.DecimalSeparator)

	writer.WriteUint64(token.MaxSupply)
	writer.WriteUint64(token.Supply)

	writer.Write(token.Key[:])
	writer.Write(token.SupplyKey[:])

	writer.WriteString(token.Name)
	writer.WriteString(token.Ticker)
	writer.WriteString(token.Description)

	return writer.Bytes()
}

func (token *Token) Deserialize(buf []byte) (err error) {

	reader := helpers.NewBufferReader(buf)

	if token.Version, err = reader.ReadUvarint(); err != nil {
		return err
	}

	if token.CanUpgrade, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.CanMint, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.CanBurn, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.CanChangeKey, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.CanChangeSupplyKey, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.CanPause, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.CanFreeze, err = reader.ReadBool(); err != nil {
		return err
	}
	if token.DecimalSeparator, err = reader.ReadByte(); err != nil {
		return err
	}
	if token.MaxSupply, err = reader.ReadUvarint(); err != nil {
		return err
	}
	if token.Supply, err = reader.ReadUvarint(); err != nil {
		return err
	}

	var data []byte
	if data, err = reader.ReadBytes(20); err != nil {
		return err
	}
	token.Key = *helpers.Byte20(data)

	if data, err = reader.ReadBytes(20); err != nil {
		return err
	}
	token.SupplyKey = *helpers.Byte20(data)

	if token.Name, err = reader.ReadString(); err != nil {
		return
	}
	if token.Ticker, err = reader.ReadString(); err != nil {
		return err
	}
	if token.Description, err = reader.ReadString(); err != nil {
		return err
	}

	return
}
