package token

import (
	"errors"
	"math"
	"pandora-pay/config/config_tokens"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"regexp"
)

var regexTokenName = regexp.MustCompile("^([a-zA-Z0-9]+ )+[a-zA-Z0-9]+$|^[a-zA-Z0-9]+")
var regexTokenTicker = regexp.MustCompile("^[A-Z0-9]+$") // only lowercase ascii is allowed. No space allowed
var regexTokenDescription = regexp.MustCompile("[\\w|\\W]+")

type Token struct {
	helpers.SerializableInterface `json:"-"`
	Version                       uint64           `json:"version,omitempty"`
	CanUpgrade                    bool             `json:"canUpgrade,omitempty"`               //upgrade different settings
	CanMint                       bool             `json:"canMint,omitempty"`                  //increase supply
	CanBurn                       bool             `json:"canBurn,omitempty"`                  //decrease supply
	CanChangePublicKey            bool             `json:"canChangePublicKey,omitempty"`       //can change key
	CanChangeSupplyPublicKey      bool             `json:"canChangeSupplyPublicKey,omitempty"` //can change supply key
	CanPause                      bool             `json:"canPause,omitempty"`                 //can pause (suspend transactions)
	CanFreeze                     bool             `json:"canFreeze,omitempty"`                //freeze supply changes
	DecimalSeparator              byte             `json:"decimalSeparator,omitempty"`
	MaxSupply                     uint64           `json:"maxSupply,omitempty"`
	Supply                        uint64           `json:"supply,omitempty"`
	UpdatePublicKey               helpers.HexBytes `json:"updatePublicKey,omitempty"` //33 byte
	SupplyPublicKey               helpers.HexBytes `json:"supplyPublicKey,omitempty"` //33 byte
	Name                          string           `json:"name"`
	Ticker                        string           `json:"ticker"`
	Description                   string           `json:"description,omitempty"`
}

func (token *Token) Validate() error {

	if token.DecimalSeparator > config_tokens.TOKENS_DECIMAL_SEPARATOR_MAX_BYTE {
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

func (token *Token) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(token.Version)

	w.WriteBool(token.CanUpgrade)
	w.WriteBool(token.CanMint)
	w.WriteBool(token.CanBurn)
	w.WriteBool(token.CanChangePublicKey)
	w.WriteBool(token.CanChangeSupplyPublicKey)
	w.WriteBool(token.CanPause)
	w.WriteBool(token.CanFreeze)
	w.WriteByte(token.DecimalSeparator)

	w.WriteUvarint(token.MaxSupply)
	w.WriteUvarint(token.Supply)

	w.Write(token.UpdatePublicKey)
	w.Write(token.SupplyPublicKey)

	w.WriteString(token.Name)
	w.WriteString(token.Ticker)
	w.WriteString(token.Description)

}

func (token *Token) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	token.Serialize(w)
	return w.Bytes()
}

func (token *Token) Deserialize(r *helpers.BufferReader) (err error) {

	if token.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if token.CanUpgrade, err = r.ReadBool(); err != nil {
		return
	}
	if token.CanMint, err = r.ReadBool(); err != nil {
		return
	}
	if token.CanBurn, err = r.ReadBool(); err != nil {
		return
	}
	if token.CanChangePublicKey, err = r.ReadBool(); err != nil {
		return
	}
	if token.CanChangeSupplyPublicKey, err = r.ReadBool(); err != nil {
		return
	}
	if token.CanPause, err = r.ReadBool(); err != nil {
		return
	}
	if token.CanFreeze, err = r.ReadBool(); err != nil {
		return
	}
	if token.DecimalSeparator, err = r.ReadByte(); err != nil {
		return
	}
	if token.MaxSupply, err = r.ReadUvarint(); err != nil {
		return
	}
	if token.Supply, err = r.ReadUvarint(); err != nil {
		return
	}
	if token.UpdatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if token.SupplyPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if token.Name, err = r.ReadString(); err != nil {
		return
	}
	if token.Ticker, err = r.ReadString(); err != nil {
		return
	}
	if token.Description, err = r.ReadString(); err != nil {
		return
	}

	return
}
