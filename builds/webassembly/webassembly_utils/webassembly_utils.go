package webassembly_utils

import (
	"encoding/json"
	"pandora-pay/recovery"
	"syscall/js"
)

var PromiseConstructor, ErrorConstructor js.Value

func ConvertToJSONBytes(reply any, err error) (js.Value, error) {

	if err != nil {
		return js.Null(), err
	}

	return ConvertJSONBytes(reply)
}

func ConvertJSONBytes(obj interface{}) (js.Value, error) {

	data, err := json.Marshal(obj)
	if err != nil {
		return js.Null(), err
	}

	return ConvertBytes(data), nil
}

func ConvertBytes(data []byte) js.Value {
	if data == nil {
		return js.Null()
	}
	jsOut := js.Global().Get("Uint8Array").New(len(data))
	js.CopyBytesToJS(jsOut, data)
	return jsOut
}

func UnmarshalBytes(data js.Value, obj interface{}) error {
	jsonData := make([]byte, data.Get("length").Int())
	js.CopyBytesToGo(jsonData, data)
	return json.Unmarshal(jsonData, obj)
}

func GetBytes(data js.Value) []byte {
	bytesArray := make([]byte, data.Get("length").Int())
	js.CopyBytesToGo(bytesArray, data)
	return bytesArray
}

func PromiseFunction(callback func() (interface{}, error)) interface{} {

	return PromiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		recovery.SafeGo(func() {
			result, err := callback()
			if err != nil {
				args2[1].Invoke(ErrorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke(result)
		})
		return nil
	}))

}

func init() {
	PromiseConstructor = js.Global().Get("Promise")
	ErrorConstructor = js.Global().Get("Error")

	if PromiseConstructor.IsNull() || ErrorConstructor.IsNull() {
		panic("promiseConstructor is null")
	}
}
