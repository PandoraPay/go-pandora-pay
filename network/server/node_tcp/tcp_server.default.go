//go:build !wasm
// +build !wasm

package node_tcp

import (
	"crypto/tls"
	"errors"
	"golang.org/x/crypto/acme/autocert"
	"net"
	"net/http"
	"net/url"
	"os"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/server/node_http"
	"pandora-pay/recovery"
	"pandora-pay/settings"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"path"
	"strconv"
	"time"
)

type TcpServer struct {
	Address     string
	Port        string
	URL         *url.URL
	tcpListener net.Listener
	HttpServer  *node_http.HttpServer
}

func NewTcpServer(connectedNodes *connected_nodes.ConnectedNodes, bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, txsValidator *txs_validator.TxsValidator, txsBuilder *txs_builder.TxsBuilder) (*TcpServer, error) {

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

	server.Port = port

	shareAddress := true
	if address == "na" {
		shareAddress = false
		address = ""
	}

	if shareAddress {
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
	}

	bannedNodes.Ban(&url.URL{Scheme: "ws", Host: "127.0.0.1:" + port, Path: "/ws"}, "", "You can't connect to yourself", 10*365*24*time.Hour)
	bannedNodes.Ban(&url.URL{Scheme: "ws", Host: address + ":" + port, Path: "/ws"}, "", "You can't connect to yourself", 10*365*24*time.Hour)

	var certPath, keyPath string
	if globals.Arguments["--tcp-server-tls-cert-file"] != nil {
		certPath = globals.Arguments["--tcp-server-tls-cert-file"].(string)
	} else {
		certPath = path.Join(config.ORIGINAL_PATH, "certificate.crt")
	}

	if globals.Arguments["--tcp-server-tls-key-file"] != nil {
		keyPath = globals.Arguments["--tcp-server-tls-key-file"].(string)
	} else {
		keyPath = path.Join(config.ORIGINAL_PATH, "certificate.key")
	}

	var tlsConfig *tls.Config
	if _, err = os.Stat(certPath); err == nil {
		cer, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cer}}
	} else if globals.Arguments["--tcp-server-auto-tls-certificate"] == true {

		if globals.Arguments["--tcp-server-address"] == "" {
			return nil, errors.New("To get an automatic Automatic you need to specify a domain --tcp-server-address=\"domain.com\"")
		}

		cache := path.Join(config.ORIGINAL_PATH, "certManager")

		if _, err = os.Stat(cache); os.IsNotExist(err) {
			if err = os.Mkdir(cache, 0755); err != nil {
				return nil, err
			}
		}

		// create the autocert.Manager with domains and path to the cache
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(address),
			Cache:      autocert.DirCache(cache), //it is designed to avoid generating multiple certificates for the same instance
		}

		tlsConfig = certManager.TLSConfig()

	}

	if shareAddress {
		websocketUrl := &url.URL{Scheme: "ws", Host: address + ":" + port, Path: "/ws"}
		url := &url.URL{Scheme: "http", Host: address + ":" + port, Path: ""}
		if tlsConfig != nil {
			websocketUrl.Scheme += "s"
			url.Scheme += "s"
		}
		config.NETWORK_ADDRESS_URL_STRING = url.String()
		config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING = websocketUrl.String()

		bannedNodes.Ban(websocketUrl, "", "You can't connect to yourself", 10*365*24*time.Hour)
		server.URL = url
	}

	if tlsConfig != nil {
		if server.tcpListener, err = tls.Listen("tcp", ":"+port, tlsConfig); err != nil {
			return nil, err
		}
		gui.GUI.Info("TLS Certificate loaded for ", address, port)
	} else {
		// no ssl at all
		if server.tcpListener, err = net.Listen("tcp", ":"+port); err != nil {
			return nil, errors.New("Error creating TcpServer" + err.Error())
		}
		gui.GUI.Warning("No TLS Certificate")
	}

	gui.GUI.InfoUpdate("TCP", address+":"+port)

	if server.HttpServer, err = node_http.NewHttpServer(chain, settings, connectedNodes, bannedNodes, knownNodes, mempool, wallet, txsValidator, txsBuilder); err != nil {
		return nil, err
	}

	recovery.SafeGo(func() {
		if err := http.Serve(server.tcpListener, *server.HttpServer.GetHttpHandler()); err != nil {
			gui.GUI.Error("Error opening HTTP server", err)
		}
		gui.GUI.Info("HTTP server")
	})

	return server, nil
}
