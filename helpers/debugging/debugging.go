package debugging

import (
	"net/http"
	_ "net/http/pprof"
)

func Start() {
	http.ListenAndServe("localhost:6060", nil)
}
