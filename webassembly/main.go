package main

import (
	"fmt"
	"pandora-pay/config"
)

func main() {
	var err error

	config.StartConfig()

	fmt.Println("PANDORA PAY WASM")

	if err = config.InitConfig(); err != nil {
	}

	fmt.Println(config.NAME)
	fmt.Println("VERSION: " + config.VERSION)
	fmt.Println("ARHITECTURE: " + config.ARCHITECTURE)
}
