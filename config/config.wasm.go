//go:build wasm
// +build wasm

package config

import (
	"time"
)

const WEBSOCKETS_TIMEOUT = 10 * time.Second //seconds

func config_init() (err error) {
	return
}
