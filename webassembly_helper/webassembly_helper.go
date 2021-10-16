package main

import (
	"os"
	"os/signal"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"syscall"
	"syscall/js"
)

func main() {

	config.StartConfig()
	argv := os.Args[1:]
	if err := arguments.InitArguments(argv); err != nil {
		panic(err)
	}
	if err := config.InitConfig(); err != nil {
		panic(err)
	}

	js.Global().Set("PandoraPay", js.ValueOf(map[string]interface{}{
		"helloPandoraHelper": js.FuncOf(helloPandoraHelper),
		"wallet": js.ValueOf(map[string]interface{}{
			"initializeBalanceDecoder": js.FuncOf(initializeBalanceDecoder),
			"decodeBalance":            js.FuncOf(decodeBalance),
		}),
		"transactions": js.ValueOf(map[string]interface{}{
			"builder": js.ValueOf(map[string]interface{}{
				"createZetherTx":              js.FuncOf(createZetherTx),
				"createZetherDelegateStakeTx": js.FuncOf(createZetherDelegateStakeTx),
				"createZetherClaimStakeTx":    js.FuncOf(createZetherClaimStakeTx),
			}),
		}),
	}))

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}
