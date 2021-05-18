package main

import (
	"fmt"
	"pandora-pay/config"
)

func main(){
	config.InitConfig("WASM")
	fmt.Println(config.NAME)
	fmt.Println("VERSION: "+config.VERSION)
	fmt.Println("ARHITECTURE: "+config.ARCHITECTURE)
}