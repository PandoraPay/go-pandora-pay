package main

import (
	"fmt"
	"os"
	"os/signal"
	"pandora-pay/blockchain"
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

	store.DBClose()

	fmt.Println("Shutting down")
}

func main() {

	gui.GUIInit()

	store.DBInit()

	wallet.WalletInit()
	settings.SettingsInit()

	blockchain.BlockchainInit()

	gui.Log("Main Loop")

	mainloop()
}
