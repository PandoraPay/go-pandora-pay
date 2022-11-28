package debugging_pprof

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"pandora-pay/config"
	"pandora-pay/helpers/recovery"
	"strconv"
)

func Start() (err error) {

	recovery.SafeGo(func() {
		addr := "localhost:" + strconv.Itoa(6060+config.INSTANCE_ID)
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
		fmt.Println("DEBUGGING STARTED ON ", addr)
	})

	return nil
}
