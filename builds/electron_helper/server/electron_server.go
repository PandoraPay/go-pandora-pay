package server

import (
	"net/http"
	"pandora-pay/builds/builds_data"
	"pandora-pay/builds/electron_helper/server/routes"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"strconv"
)

func CreateServer() error {

	mux := http.NewServeMux()

	mux.HandleFunc("/wallet/initialize-balance-decryptor", serverMethod[builds_data.WalletInitializeBalanceDecryptorReq](routes.RouteWalletInitializeBalanceDecryptor))
	mux.HandleFunc("/wallet/decrypt-balance", serverMethod[builds_data.WalletDecryptBalanceReq](routes.RouteWalletDecryptBalance))
	mux.HandleFunc("/transactions/builder/create-zether-transaction", serverMethod[builds_data.TransactionsBuilderCreateZetherTxReq](routes.RouteTransactionsBuilderCreateZetherTx))
	mux.HandleFunc("/", routes.RouteHome)

	port := globals.Arguments["--tcp-server-port"].(string)

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	//Use the default DefaultServeMux.
	gui.GUI.Log("Opening a HTTP server :", portNumber)
	go func() {

		err := http.ListenAndServe(":"+strconv.Itoa(portNumber), mux)
		gui.GUI.Log("HTTP server open :", portNumber)

		if err != nil {
			panic(err)
		}
	}()

	return nil
}
