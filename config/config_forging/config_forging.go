package config_forging

import "pandora-pay/config/globals"

var (
	FORGING_ENABLED = true
)

func InitConfig() (err error) {

	if globals.Arguments["--forging"] == false {
		FORGING_ENABLED = false
	}

	return
}
