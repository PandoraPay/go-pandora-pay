// +build wasm

package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/events"
	"sync/atomic"
	"syscall/js"
)

var subscriptionsIndex uint64
var startMainCallback func()

func SubscribeEvents(none js.Value, args []js.Value) interface{} {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel := globals.MainEvents.AddListener()
	callback := args[0]

	go func() {
		for {
			dataValue := <-channel
			data := dataValue.(*events.EventData)

			var final interface{}

			switch v := data.Data.(type) {
			case interface{}:
				str, err := json.Marshal(v)
				if err == nil {
					final = string(str)
				} else {
					final = "error marshaling object"
				}
			default:
				final = data.Data
			}

			callback.Invoke(data.Name, final)
		}
	}()

	return index
}

func HelloPandora(js.Value, []js.Value) interface{} {
	gui.GUI.Info("HelloPandora works!")
	return true
}

func Start(js.Value, []js.Value) interface{} {
	startMainCallback()
	return true
}

func Initialize(startMainCb func()) {

	startMainCallback = startMainCb

	Events := map[string]interface{}{
		"Subscribe": js.FuncOf(SubscribeEvents),
	}

	Helpers := map[string]interface{}{
		"HelloPandora": js.FuncOf(HelloPandora),
		"Start":        js.FuncOf(Start),
	}

	PandoraPayExport := map[string]interface{}{
		"Helpers": js.ValueOf(Helpers),
		"Events":  js.ValueOf(Events),
	}

	js.Global().Set("PandoraPay", js.ValueOf(PandoraPayExport))

}
