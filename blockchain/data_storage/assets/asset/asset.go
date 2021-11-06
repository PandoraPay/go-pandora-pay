package asset

import (
	"bytes"
	"errors"
	"math"
	"pandora-pay/config/config_assets"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"regexp"
)

var regexAssetName = regexp.MustCompile("^([a-zA-Z0-9]+ )+[a-zA-Z0-9]+$|^[a-zA-Z0-9]+")
var regexAssetTicker = regexp.MustCompile("^[A-Z0-9]+$") // only lowercase ascii is allowed. No space allowed
var regexAssetDescription = regexp.MustCompile("[\\w|\\W]+")

type Asset struct {
	helpers.SerializableInterface `json:"-"`
	Version                       uint64           `json:"version,omitempty"`
	CanUpgrade                    bool             `json:"canUpgrade,omitempty"`               //upgrade different settings
	CanMint                       bool             `json:"canMint,omitempty"`                  //increase supply
	CanBurn                       bool             `json:"canBurn,omitempty"`                  //decrease supply
	CanChangeUpdatePublicKey      bool             `json:"canChangeUpdatePublicKey,omitempty"` //can change key
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

func (asset *Asset) Validate() error {

	if asset.DecimalSeparator > config_assets.ASSETS_DECIMAL_SEPARATOR_MAX_BYTE {
		return errors.New("asset decimal separator is invalid")
	}
	if len(asset.Name) > 15 || len(asset.Name) < 3 {
		return errors.New("asset name length is invalid")
	}
	if len(asset.Ticker) > 10 || len(asset.Ticker) < 2 {
		return errors.New("asset ticker length is invalid")
	}
	if len(asset.Description) > 512 {
		return errors.New("asset  description length is invalid")
	}

	if !regexAssetName.MatchString(asset.Name) {
		return errors.New("Asset name is invalid")
	}
	if !regexAssetTicker.MatchString(asset.Ticker) {
		return errors.New("Asset ticker is invalid")
	}
	if !regexAssetDescription.MatchString(asset.Description) {
		return errors.New("Asset description is invalid")
	}

	return nil
}

func (asset *Asset) ConvertToUnits(amount float64) (uint64, error) {
	COIN_DENOMINATION := math.Pow10(int(asset.DecimalSeparator))
	if amount < float64(math.MaxUint64)/COIN_DENOMINATION {
		return uint64(amount * COIN_DENOMINATION), nil
	}
	return 0, errors.New("Error converting to units")
}

func (asset *Asset) ConvertToBase(amount uint64) float64 {
	COIN_DENOMINATION := math.Pow10(int(asset.DecimalSeparator))
	return float64(amount) / COIN_DENOMINATION
}

func (asset *Asset) AddSupply(sign bool, amount uint64, blockRewards bool) error {

	if !blockRewards && bytes.Equal(asset.SupplyPublicKey, config_coins.BURN_PUBLIC_KEY) {
		return errors.New("BURN PUBLIC KEY")
	}

	if sign {
		if !asset.CanMint && !blockRewards {
			return errors.New("Can't mint")
		}
		if asset.MaxSupply-asset.Supply < amount {
			return errors.New("Supply exceeded max supply")
		}
		return helpers.SafeUint64Add(&asset.Supply, amount)
	}

	if !asset.CanBurn && !blockRewards {
		return errors.New("Can't burn")
	}
	if asset.Supply < amount {
		return errors.New("Supply would become negative")
	}
	return helpers.SafeUint64Sub(&asset.Supply, amount)
}

func (asset *Asset) Serialize(w *helpers.BufferWriter) {

	w.WriteUvarint(asset.Version)

	w.WriteBool(asset.CanUpgrade)
	w.WriteBool(asset.CanMint)
	w.WriteBool(asset.CanBurn)
	w.WriteBool(asset.CanChangeUpdatePublicKey)
	w.WriteBool(asset.CanChangeSupplyPublicKey)
	w.WriteBool(asset.CanPause)
	w.WriteBool(asset.CanFreeze)
	w.WriteByte(asset.DecimalSeparator)

	w.WriteUvarint(asset.MaxSupply)
	w.WriteUvarint(asset.Supply)

	w.Write(asset.UpdatePublicKey)
	w.Write(asset.SupplyPublicKey)

	w.WriteString(asset.Name)
	w.WriteString(asset.Ticker)
	w.WriteString(asset.Description)

}

func (asset *Asset) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	asset.Serialize(w)
	return w.Bytes()
}

func (asset *Asset) Deserialize(r *helpers.BufferReader) (err error) {

	if asset.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if asset.CanUpgrade, err = r.ReadBool(); err != nil {
		return
	}
	if asset.CanMint, err = r.ReadBool(); err != nil {
		return
	}
	if asset.CanBurn, err = r.ReadBool(); err != nil {
		return
	}
	if asset.CanChangeUpdatePublicKey, err = r.ReadBool(); err != nil {
		return
	}
	if asset.CanChangeSupplyPublicKey, err = r.ReadBool(); err != nil {
		return
	}
	if asset.CanPause, err = r.ReadBool(); err != nil {
		return
	}
	if asset.CanFreeze, err = r.ReadBool(); err != nil {
		return
	}
	if asset.DecimalSeparator, err = r.ReadByte(); err != nil {
		return
	}
	if asset.MaxSupply, err = r.ReadUvarint(); err != nil {
		return
	}
	if asset.Supply, err = r.ReadUvarint(); err != nil {
		return
	}
	if asset.UpdatePublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if asset.SupplyPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if asset.Name, err = r.ReadString(); err != nil {
		return
	}
	if asset.Ticker, err = r.ReadString(); err != nil {
		return
	}
	if asset.Description, err = r.ReadString(); err != nil {
		return
	}

	return
}
