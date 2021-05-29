package gui_interface

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/addresses"
	"pandora-pay/helpers"
	"strconv"
)

type GUIInterface interface {
	Close()
	Log(any ...interface{})
	Info(any ...interface{})
	Warning(any ...interface{})
	Fatal(any ...interface{})
	Error(any ...interface{})
	InfoUpdate(key string, text string)
	Info2Update(key string, text string)
	OutputWrite(any interface{})
	CommandDefineCallback(Text string, callback func(string) error)
	OutputReadString(text string) (out string, ok bool)
	OutputReadInt(text string, acceptedValues []int) (out int, ok bool)
	OutputReadUint64(text string, acceptedValues []uint64, acceptEmpty bool) (out uint64, ok bool)
	OutputReadFloat64(text string, acceptedValues []float64) (out float64, ok bool)
	OutputReadAddress(text string) (address *addresses.Address, ok bool)
	OutputReadBool(text string) (out bool, ok bool)
	OutputReadBytes(text string, acceptedLengths []int) (token []byte, ok bool)
}

func ProcessArgument(any ...interface{}) string {

	var s = ""

	for i, it := range any {

		if i > 0 {
			s += " "
		}

		switch v := it.(type) {
		case nil:
			s += "nil"
		case bool:
			s += strconv.FormatBool(v)
		case string:
			s += v
		case int:
			s += strconv.Itoa(v)
		case float64:
			s += strconv.FormatFloat(v, 'f', 10, 64)
		case uint64:
			s += strconv.FormatUint(v, 10)
		case []byte:
		case helpers.HexBytes:
			s += hex.EncodeToString(v)
		case error:
			s += v.Error()
		case interface{}:
			str, err := json.Marshal(v)
			if err == nil {
				s += string(str)
			} else {
				s += "error marshaling object"
			}
		default:
			s += "invalid log type"
		}

	}

	return s
}
