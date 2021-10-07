package main

import (
	"os"
	"os/signal"
	"syscall"
	"syscall/js"
)

func main() {

	js.Global().Set("PandoraPay", js.ValueOf(map[string]interface{}{
		"helloPandoraHelper": js.FuncOf(helloPandoraHelper),
		"wallet": js.ValueOf(map[string]interface{}{
			"decodeBalance": js.FuncOf(decodeBalance),
		}),
	}))

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}
