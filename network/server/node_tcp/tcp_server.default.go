//go:build !wasm
// +build !wasm

package node_tcp

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/server/node_http"
	"pandora-pay/recovery"
	"pandora-pay/settings"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
	"strconv"
	"time"
)

// ControllerAddr is the Tor controller interface address
// Note:
type TcpServer struct {
	Address     string
	Port        string
	URL         *url.URL
	tcpListener net.Listener
	HttpServer  *node_http.HttpServer
}

func CreateTcpServer(bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*TcpServer, error) {

	server := &TcpServer{}

	// Create local listener on next available port

	port := globals.Arguments["--tcp-server-port"].(string)

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.New("Port is not a valid port number")
	}

	portNumber += config.INSTANCE_ID

	port = strconv.Itoa(portNumber)

	var address string
	if globals.Arguments["--tor-onion"] != nil {
		address = globals.Arguments["--tor-onion"].(string)
	}
	if globals.Arguments["--tcp-server-address"] != nil {
		address = globals.Arguments["--tcp-server-address"].(string)
	}

	if address == "" {
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			return nil, errors.New("Error dialing dns to discover my own ip" + err.Error())
		}
		address = conn.LocalAddr().(*net.UDPAddr).IP.String()
		if err = conn.Close(); err != nil {
			return nil, errors.New("Error closing the connection" + err.Error())
		}
	}
	server.Address = address
	server.Port = port
	server.URL = &url.URL{Scheme: "ws", Host: address + ":" + port, Path: "/ws"}

	config.NETWORK_ADDRESS_URL = server.URL
	config.NETWORK_ADDRESS_URL_STRING = server.URL.String()

	bannedNodes.Ban(server.URL, "", "You can't connect to yourself", 10*365*24*time.Hour)
	bannedNodes.Ban(&url.URL{Scheme: "ws", Host: "127.0.0.1:" + port, Path: "/ws"}, "", "You can't connect to yourself", 10*365*24*time.Hour)

	server.tcpListener, err = net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, errors.New("Error creating TcpServer" + err.Error())
	}

	gui.GUI.InfoUpdate("TCP", address+":"+port)

	if server.HttpServer, err = node_http.CreateHttpServer(chain, settings, bannedNodes, knownNodes, mempool, wallet, transactionsBuilder); err != nil {
		return nil, err
	}

	recovery.SafeGo(func() {
		if err := http.Serve(server.tcpListener, nil); err != nil {
			gui.GUI.Error("Error opening HTTP server", err)
		}
		gui.GUI.Info("HTTP server")
	})

	return server, nil
}
