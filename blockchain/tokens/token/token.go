package token

import (
	"errors"
	"math"
	"pandora-pay/helpers"
	"regexp"
)

var regexTokenName = regexp.MustCompile("^([a-zA-Z0-9]+ )+[a-zA-Z0-9]+$|^[a-zA-Z0-9]+")
var regexTokenTicker = regexp.MustCompile("^[A-Z0-9]+$") // only lowercase ascii is allowed. No space allowed
var regexTokenDescription = regexp.MustCompile("[\\w|\\W]+")

type Token struct {
	helpers.SerializableInterface
	Version            uint64
	CanUpgrade         bool //upgrade different settings
	CanMint            bool //increase supply
	CanBurn            bool //decrease supply
	CanChangeKey       bool //can change key
	CanChangeSupplyKey bool //can change supply key
	CanPause           bool //can pause (suspend transactions)
	CanFreeze          bool //freeze supply changes
	DecimalSeparator   byte
	MaxSupply          uint64
	Supply             uint64
	Key                helpers.HexBytes //20 byte
	SupplyKey          helpers.HexBytes //20 byte
	Name               string
	Ticker             string
	Description        string
}

func (token *Token) Validate() error {

	if token.DecimalSeparator > 10 {
		return errors.New("token decimal separator is invalid")
	}
	if len(token.Name) > 15 || len(token.Name) < 3 {
		return errors.New("token name length is invalid")
	}
	if len(token.Ticker) > 7 || len(token.Ticker) < 2 {
		return errors.New("token ticker length is invalid")
	}
	if len(token.Description) > 512 {
		return errors.New("token  description length is invalid")
	}

	if !regexTokenName.MatchString(token.Name) {
		return errors.New("Token name is invalid")
	}
	if !regexTokenTicker.MatchString(token.Ticker) {
		return errors.New("Token ticker is invalid")
	}
	if !regexTokenDescription.MatchString(token.Description) {
		return errors.New("Token description is invalid")
	}

	return nil
}

func (token *Token) ConvertToUnits(amount float64) (uint64, error) {
	COIN_DENOMINATION := math.Pow10(int(token.DecimalSeparator))
	if amount < float64(math.MaxUint64)/COIN_DENOMINATION {
		return uint64(amount * COIN_DENOMINATION), nil
	}
	return 0, errors.New("Error converting to units")
}

func (token *Token) ConvertToBase(amount uint64) float64 {
	COIN_DENOMINATION := math.Pow10(int(token.DecimalSeparator))
	return float64(amount) / COIN_DENOMINATION
}

func (token *Token) AddSupply(sign bool, amount uint64) error {

	if sign {
		if !token.CanMint {
			return errors.New("Can't mint")
		}
		if token.MaxSupply-token.Supply < amount {
			return errors.New("Supply exceeded max supply")
		}
		return helpers.SafeUint64Add(&token.Supply, amount)
	}

	if !token.CanBurn {
		errors.New("Can't burn")
	}
	if token.Supply < amount {
		errors.New("Supply would become negative")
	}
	return helpers.SafeUint64Sub(&token.Supply, amount)
}

func (token *Token) Serialize(writer *helpers.BufferWriter) {

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

	writer.Write(token.Key)
	writer.Write(token.SupplyKey)

	writer.WriteString(token.Name)
	writer.WriteString(token.Ticker)
	writer.WriteString(token.Description)

}

func (token *Token) SerializeToBytes() []byte {
	writer := helpers.NewBufferWriter()
	token.Serialize(writer)
	return writer.Bytes()
}

func (token *Token) Deserialize(reader *helpers.BufferReader) (err error) {

	if token.Version, err = reader.ReadUvarint(); err != nil {
		return
	}
	if token.CanUpgrade, err = reader.ReadBool(); err != nil {
		return
	}
	if token.CanMint, err = reader.ReadBool(); err != nil {
		return
	}
	if token.CanBurn, err = reader.ReadBool(); err != nil {
		return
	}
	if token.CanChangeKey, err = reader.ReadBool(); err != nil {
		return
	}
	if token.CanChangeSupplyKey, err = reader.ReadBool(); err != nil {
		return
	}
	if token.CanPause, err = reader.ReadBool(); err != nil {
		return
	}
	if token.CanFreeze, err = reader.ReadBool(); err != nil {
		return
	}
	if token.DecimalSeparator, err = reader.ReadByte(); err != nil {
		return
	}
	if token.MaxSupply, err = reader.ReadUvarint(); err != nil {
		return
	}
	if token.Supply, err = reader.ReadUvarint(); err != nil {
		return
	}
	if token.Key, err = reader.ReadBytes(20); err != nil {
		return
	}
	if token.SupplyKey, err = reader.ReadBytes(20); err != nil {
		return
	}
	if token.Name, err = reader.ReadString(); err != nil {
		return
	}
	if token.Ticker, err = reader.ReadString(); err != nil {
		return
	}
	if token.Description, err = reader.ReadString(); err != nil {
		return
	}

	return
}
