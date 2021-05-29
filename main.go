package main

import (
	"fmt"
	"os"
	"os/signal"
	"pandora-pay/config/globals"
	"pandora-pay/helpers/events"
	"pandora-pay/start"
	"syscall"
)

func main() {

	globals.MainEvents = events.NewEvents()

	start.RunMain()

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	start.CloseMain()

	fmt.Println("Shutting down")
}
