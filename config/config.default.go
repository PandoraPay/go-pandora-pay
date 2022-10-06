//go:build !wasm
// +build !wasm

package config

import (
	"os"
	"pandora-pay/config/globals"
	"strconv"
	"time"
)

const WEBSOCKETS_TIMEOUT = 5 * time.Second //seconds

func config_init() (err error) {

	if _, err = os.Stat("./_build"); os.IsNotExist(err) {
		if err = os.Mkdir("./_build", 0755); err != nil {
			return
		}
	}

	if ORIGINAL_PATH, err = os.Getwd(); err != nil {
		return
	}

	if err = os.Chdir("./_build"); err != nil {
		return
	}

	var prefix string
	if globals.Arguments["--instance"] != nil {
		INSTANCE = globals.Arguments["--instance"].(string)
		prefix = INSTANCE
	} else {
		prefix = "default"
	}

	if globals.Arguments["--instance-id"] != nil {
		a := globals.Arguments["--instance-id"].(string)
		if INSTANCE_ID, err = strconv.Atoi(a); err != nil {
			return
		}
	}
	prefix += "_" + strconv.Itoa(INSTANCE_ID)

	if _, err = os.Stat("./" + prefix); os.IsNotExist(err) {
		if err = os.Mkdir("./"+prefix, 0755); err != nil {
			return
		}
	}

	prefix += "/" + NETWORK_SELECTED_NAME
	if _, err = os.Stat("./" + prefix); os.IsNotExist(err) {
		if err = os.Mkdir("./"+prefix, 0755); err != nil {
			return
		}
	}

	if err = os.Chdir("./" + prefix); err != nil {
		return
	}

	return
}
