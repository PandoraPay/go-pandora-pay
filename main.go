package main

import (
	"fmt"
	"os"
	"os/signal"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/config/globals"
	"pandora-pay/helpers/events"
	"pandora-pay/recovery"
	"pandora-pay/start"
	"runtime"
	"syscall"
)

func saveError(err error) {

	fmt.Println(err)

	if runtime.GOARCH != "wasm" {
		return
	}

	file, err2 := os.Create("./error.txt")
	if err2 != nil {
		panic(err2)
	}
	defer file.Close()
	if _, err2 = file.Write([]byte(err.Error())); err2 != nil {
		panic(err2)
	}

}

func main() {

	var err error
	globals.MainEvents = events.NewEvents[any]()

	config.StartConfig()

	argv := os.Args[1:]
	if err = arguments.InitArguments(argv); err != nil {
		saveError(err)
		panic(err)
	}

	if err = config.InitConfig(); err != nil {
		saveError(err)
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "config initialized")

	recovery.Safe(start.RunMain)

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	fmt.Println("Shutting down")

}
