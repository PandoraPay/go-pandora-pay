package api_code_http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api_code/api_code_types"
)

func HandleAuthenticated[T any, B any](callback func(r *http.Request, args *T, reply *B, authenticated bool) error) func(values url.Values) (interface{}, error) {
	return func(values url.Values) (interface{}, error) {

		authenticated := api_code_types.CheckAuthenticated(values)
		values.Del("user")
		values.Del("pass")

		args := new(T)
		if err := urldecoder.Decoder.Decode(args, values); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply, authenticated)
	}
}

func Handle[T any, B any](callback func(r *http.Request, args *T, reply *B) error) func(values url.Values) (interface{}, error) {
	return func(values url.Values) (interface{}, error) {
		args := new(T)
		if err := urldecoder.Decoder.Decode(args, values); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply)
	}
}

func HandlePOSTAuthenticated[T any, B any](callback func(r *http.Request, args *T, reply *B, authenticated bool) error) func(values io.ReadCloser) (interface{}, error) {
	return func(values io.ReadCloser) (interface{}, error) {

		authenticated := new(api_code_types.APIAuthenticated[T])
		if err := json.NewDecoder(values).Decode(authenticated); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, authenticated.Data, reply, authenticated.CheckAuthenticated())
	}
}

func HandlePOST[T any, B any](callback func(r *http.Request, args *T, reply *B) error) func(values io.ReadCloser) (interface{}, error) {
	return func(values io.ReadCloser) (interface{}, error) {
		args := new(T)

		if err := json.NewDecoder(values).Decode(args); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply)
	}
}
