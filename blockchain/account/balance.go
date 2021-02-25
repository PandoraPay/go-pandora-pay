package account

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type Balance struct {
	Amount   uint64
	Currency [20]byte
}

func (balance *Balance) Serialize(serialized *bytes.Buffer, buf []byte) {

	n := binary.PutUvarint(buf, balance.Amount)
	serialized.Write(buf[:n])

	serialized.Write(balance.Currency[:])

}

func (balance *Balance) Deserialize(buf []byte) (out []byte, err error) {

	balance.Amount, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	var data []byte
	data, buf, err = helpers.DeserializeBuffer(buf, 20)
	if err != nil {
		return
	}
	copy(balance.Currency[:], data)

	out = buf
	return
}
