//go:build js
// +build js

package start

import (
	"fmt"
)

func saveError(err error) {
	fmt.Println(err)
	panic(err)
}

func startMain() {

}
