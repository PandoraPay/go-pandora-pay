package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func serverMethod[T any](method func(*T) (any, error)) func(http.ResponseWriter, *http.Request) {

	f := func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")

		defer func() {
			if err := recover(); err != nil {
				http.Error(w, err.(error).Error(), http.StatusInternalServerError)
			}
		}()

		value := new(T)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		if err = json.Unmarshal(body, value); err != nil {
			panic(err)
		}

		result, err := method(value)
		if err != nil {
			panic(err)
		}

		final, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(final)
	}

	return f
}
