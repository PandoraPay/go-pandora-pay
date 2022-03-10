package main

import (
	"os"
	"os/signal"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/gui"
	"syscall"
	"syscall/js"
)

var AddressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor

func main() {
	var err error

	config.StartConfig()
	argv := os.Args[1:]
	if err := arguments.InitArguments(argv); err != nil {
		panic(err)
	}
	if err := config.InitConfig(); err != nil {
		panic(err)
	}

	if err := gui.InitGUI(); err != nil {
		panic(err)
	}

	if AddressBalanceDecryptor, err = address_balance_decryptor.NewAddressBalanceDecryptor(); err != nil {
		return
	}

	js.Global().Set("PandoraPay", js.ValueOf(map[string]interface{}{
		"helloPandoraHelper": js.FuncOf(helloPandoraHelper),
		"wallet": js.ValueOf(map[string]interface{}{
			"initializeBalanceDecryptor": js.FuncOf(initializeBalanceDecryptor),
			"decryptBalance":             js.FuncOf(decryptBalance),
		}),
		"transactions": js.ValueOf(map[string]interface{}{
			"builder": js.ValueOf(map[string]interface{}{
				"createZetherTx": js.FuncOf(createZetherTx),
			}),
		}),
	}))

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}
