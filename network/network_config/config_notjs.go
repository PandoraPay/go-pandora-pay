//go:build !wasm
// +build !wasm

package network_config

import "time"

const WEBSOCKETS_TIMEOUT = 5 * time.Second //seconds
