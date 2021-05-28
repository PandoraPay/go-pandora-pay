// +build wasm

package webassembly

import (
	"errors"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/events"
	"sync/atomic"
	"syscall/js"
	"time"
)

var subscriptionsIndex uint64

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
			callback.Invoke(data.Name, data.Data)
		}
	}()

	return index
}

func HelloPandora(js.Value, []js.Value) interface{} {
	gui.GUI.Info("HelloPandora works!")
	return nil
}

func Initialize() {

	Events := map[string]interface{}{
		"Subscribe": js.FuncOf(SubscribeEvents),
	}

	Helpers := map[string]interface{}{
		"HelloPandora": js.FuncOf(HelloPandora),
	}

	PandoraPayExport := map[string]interface{}{
		"Helpers": js.ValueOf(Helpers),
		"Events":  js.ValueOf(Events),
	}

	js.Global().Set("PandoraPay", js.ValueOf(PandoraPayExport))

	go func() {
		for {
			globals.MainEvents.BroadcastEvent("main", "working...")
			time.Sleep(1 * time.Second)
		}
	}()

}
