package gui_interface

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
)

var GUIInterfaceError = errors.New("Ctrl+C Suspended")

type GUIInterface interface {
	Close()
	Log(any ...interface{})
	Info(any ...interface{})
	Warning(any ...interface{})
	Fatal(any ...interface{})
	Error(any ...interface{})
	InfoUpdate(key string, text string)
	Info2Update(key string, text string)
	OutputWrite(any ...interface{})
	CommandDefineCallback(Text string, callback func(string, context.Context) error, useIt bool)
	OutputReadString(text string) string
	OutputReadFilename(text, extension string, allowEmpty bool) string
	OutputReadInt(text string, allowEmpty bool, emptyValue int, validateCb func(int) bool) int
	OutputReadUint64(text string, allowEmpty bool, emptyValue uint64, validateCb func(uint64) bool) uint64
	OutputReadFloat64(text string, allowEmpty bool, emptyValue float64, validateCb func(float64) bool) float64
	OutputReadBool(text string, allowEmpty bool, emptyValue bool) (out bool)
	OutputReadBytes(text string, validateCb func([]byte) bool) (data []byte)
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
			s += base64.StdEncoding.EncodeToString(v)
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
