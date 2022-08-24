package routes

import (
	"encoding/json"
	"market/config"
	"net/http"
)

func RouteHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u := &struct {
		ProtocolVersion uint64 `json:"protocolVersion"`
		Version         string `json:"version"`
	}{
		config.PROTOCOL_VERSION,
		config.VERSION,
	}
	json.NewEncoder(w).Encode(u)
}
