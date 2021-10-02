package debugging

import (
	"net/http"
	_ "net/http/pprof"
	"pandora-pay/config"
	"pandora-pay/recovery"
	"strconv"
)

func Start() (err error) {

	recovery.SafeGo(func() {
		if err := http.ListenAndServe("localhost:"+strconv.Itoa(6060+config.INSTANCE_ID), nil); err != nil {
			panic(err)
		}
	})

	return nil
}
