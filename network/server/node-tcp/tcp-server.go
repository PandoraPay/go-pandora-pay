package node_tcp

import (
	"errors"
	"net"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
	"pandora-pay/settings"
	"strconv"
)

// ControllerAddr is the Tor controller interface address
// Note:
type TcpServer struct {
	Address     string
	Port        string
	tcpListener net.Listener
	HttpServer  *node_http.HttpServer
}

func CreateTcpServer(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) (server *TcpServer, err error) {

	server = &TcpServer{}

	// Create local listener on next available port

	port := "8080"
	if globals.Arguments["--tcp-server-port"] != nil {
		port = globals.Arguments["--tcp-server-port"].(string)
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return nil, errors.New("Port is not a valid port number")
	}

	if config.INSTANCE != "" {
		portNumber += config.INSTANCE_NUMBER
		port = strconv.Itoa(portNumber)
	}

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

	server.tcpListener, err = net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return nil, errors.New("Error creating TcpServer" + err.Error())
	}

	gui.InfoUpdate("TCP", address+":"+port)

	server.HttpServer, err = node_http.CreateHttpServer(server.tcpListener, chain, settings, mempool)

	return
}
