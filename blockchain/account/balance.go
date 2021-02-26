package account

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type Balance struct {
	Amount   uint64
	Currency []byte
}

func (balance *Balance) Serialize(serialized *bytes.Buffer, temp []byte) {

	n := binary.PutUvarint(temp, balance.Amount)
	serialized.Write(temp[:n])

	if len(balance.Currency) == 0 {
		serialized.Write([]byte{0})
	} else {
		serialized.Write([]byte{1})
		serialized.Write(balance.Currency[:])
	}

	serialized.Write(temp[:1])

}

func (balance *Balance) Deserialize(buf []byte) (out []byte, err error) {

	if balance.Amount, buf, err = helpers.DeserializeNumber(buf); err != nil {
		return
	}

	var currencyType []byte
	if currencyType, buf, err = helpers.DeserializeBuffer(buf, 1); err != nil {
		return
	}

	if currencyType[0] == 0 {
		balance.Currency = []byte{}
	} else {
		if balance.Currency, buf, err = helpers.DeserializeBuffer(buf, 20); err != nil {
			return
		}
	}

	out = buf
	return
}
