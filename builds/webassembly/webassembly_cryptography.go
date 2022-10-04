package main

import (
	"encoding/base64"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"syscall/js"
)

func sha3(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		message, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		out := cryptography.SHA3(message)

		return base64.StdEncoding.EncodeToString(out), nil
	})
}

func ripemd(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		message, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		out := cryptography.RIPEMD(message)

		return base64.StdEncoding.EncodeToString(out), nil
	})
}

func sign(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		message, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		key, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		out, err := crypto.SignMessage(message, key)
		if err != nil {
			return nil, err
		}

		return base64.StdEncoding.EncodeToString(out), nil
	})
}

func verify(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		message, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		signature, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		publicKey, err := base64.StdEncoding.DecodeString(args[2].String())
		if err != nil {
			return nil, err
		}

		out := crypto.VerifySignature(message, signature, publicKey)

		return out, nil
	})
}
