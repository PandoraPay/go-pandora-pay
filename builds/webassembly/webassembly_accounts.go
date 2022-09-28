package main

import (
	"encoding/base64"
	"pandora-pay/addresses"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"syscall/js"
)

type AddressGenerateArgument struct {
	PaymentID     []byte `json:"paymentID,omitempty"`
	PaymentAmount uint64 `json:"paymentAmount,omitempty"`
	PaymentAsset  []byte `json:"paymentAsset,omitempty"`
}

func decodeAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		address, err := addresses.DecodeAddr(args[0].String())
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(address)
	})
}

func createAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		parameters := struct {
			PublicKeyHash []byte `json:"publicKeyHash"`
			PaymentID     []byte `json:"paymentID"`
			PaymentAmount uint64 `json:"paymentAmount"`
			PaymentAsset  []byte `json:"paymentAsset"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], &parameters); err != nil {
			return nil, err
		}

		addr, err := addresses.CreateAddr(parameters.PublicKeyHash, parameters.PaymentID, parameters.PaymentAmount, parameters.PaymentAsset)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes([]interface{}{
			addr,
			addr.EncodeAddr(),
		})

	})
}

func getPublicKeyHash(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		publicKey, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		return cryptography.GetPublicKeyHash(publicKey), nil
	})
}

func generateNewAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		priv := addresses.GenerateNewPrivateKey()

		parameters := &AddressGenerateArgument{}
		if err := webassembly_utils.UnmarshalBytes(args[0], &parameters); err != nil {
			return nil, err
		}

		addr, err := priv.GenerateAddress(parameters.PaymentID, parameters.PaymentAmount, parameters.PaymentAsset)

		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes([]interface{}{
			base64.StdEncoding.EncodeToString(priv.Key),
			addr.EncodeAddr(),
			base64.StdEncoding.EncodeToString(addr.PublicKeyHash),
		})
	})
}

func generateAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		privateKey, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		priv, err := addresses.NewPrivateKey(privateKey)
		if err != nil {
			return nil, err
		}

		parameters := &AddressGenerateArgument{}
		if err := webassembly_utils.UnmarshalBytes(args[0], &parameters); err != nil {
			return nil, err
		}

		addr, err := priv.GenerateAddress(parameters.PaymentID, parameters.PaymentAmount, parameters.PaymentAsset)

		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes([]interface{}{
			base64.StdEncoding.EncodeToString(priv.Key),
			addr.EncodeAddr(),
			base64.StdEncoding.EncodeToString(addr.PublicKeyHash),
		})
	})
}
