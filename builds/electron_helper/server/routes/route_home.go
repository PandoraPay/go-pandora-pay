package routes

import (
	"encoding/json"
	"net/http"
	"pandora-pay/config"
)

func RouteHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := &struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Network uint64 `json:"network"`
	}{
		config.NAME,
		config.VERSION_STRING,
		config.NETWORK_SELECTED,
	}
	json.NewEncoder(w).Encode(u)
}
