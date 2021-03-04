package token

import (
	"errors"
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

func (token *Token) AddSupply(sign bool, amount uint64) error {

	if sign {
		if !token.CanMint {
			return errors.New("Can't mint")
		}
		if token.Supply+amount > token.MaxSupply {
			return errors.New("Supply exceeded max supply")
		}
		token.Supply += amount
	} else {
		if !token.CanBurn {
			return errors.New("Can't burn")
		}
		if token.Supply < amount {
			return errors.New("Supply would become negative")
		}

		token.Supply -= amount
	}

	return nil
}

func (token *Token) Serialize() []byte {

	writer := helpers.NewBufferWriter()

	writer.WriteUvarint(token.Version)

	writer.WriteBool(token.CanUpgrade)
	writer.WriteBool(token.CanMint)
	writer.WriteBool(token.CanBurn)
	writer.WriteBool(token.CanChangeKey)
	writer.WriteBool(token.CanChangeSupplyKey)
	writer.WriteBool(token.CanPause)
	writer.WriteBool(token.CanFreeze)
	writer.WriteByte(token.DecimalSeparator)

	writer.WriteUvarint(token.MaxSupply)
	writer.WriteUvarint(token.Supply)

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
	if token.DecimalSeparator > 10 {
		return errors.New("token decimal separator is invalid")
	}
	if token.MaxSupply, err = reader.ReadUvarint(); err != nil {
		return err
	}
	if token.Supply, err = reader.ReadUvarint(); err != nil {
		return err
	}

	if token.Key, err = reader.Read20(); err != nil {
		return err
	}

	if token.SupplyKey, err = reader.Read20(); err != nil {
		return err
	}

	if token.Name, err = reader.ReadString(); err != nil {
		return
	}
	if len(token.Name) > 15 || len(token.Name) < 3 {
		return errors.New("token name length is invalid")
	}

	if token.Ticker, err = reader.ReadString(); err != nil {
		return err
	}
	if len(token.Ticker) > 7 || len(token.Ticker) < 2 {
		return errors.New("token ticker length is invalid")
	}

	if token.Description, err = reader.ReadString(); err != nil {
		return err
	}
	if len(token.Description) > 512 {
		return errors.New("token  description length is invalid")
	}

	return
}
