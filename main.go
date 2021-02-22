package main

import (
	"fmt"
	"os"
	"os/signal"
	"pandora-pay/gui"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/wallet"
	"syscall"
)

func mainloop() {
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	store.CloseDB()

	fmt.Println("Shutting down")
}

func main() {

	gui.InitGUI()

	store.InitDB()

	wallet.InitWallet()
	settings.InitSettings()

	gui.Log("Main Loop")

	mainloop()
}
