package addresses

import (
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	base58 "pandora-pay/helpers/base58"
)

type AddressVersion uint64

const (
	TransparentPublicKeyHash AddressVersion = 0
	TransparentPublicKey     AddressVersion = 1
)

type Address struct {
	Network   uint64
	Version   uint64
	PublicKey []byte // publicKey or PublicKeyHash
	Amount    uint64 // amount to be paid
	PaymentID []byte // payment id
}

func (a *Address) EncodeAddr() (string, error) {

	var prefix string
	if a.Network == blockchain.MAIN_NET_NETWORK_BYTE {
		prefix = blockchain.MAIN_NET_NETWORK_BYTE_PREFIX
	} else if a.Network == blockchain.TEST_NET_NETWORK_BYTE {
		prefix = blockchain.TEST_NET_NETWORK_BYTE_PREFIX
	} else {
		return "", errors.New("Invalid network")
	}

	bytes := append([]byte{}, helpers.SerializeNumber(a.Version)...)
	bytes = append(bytes, a.PublicKey...)
	bytes = append(bytes, helpers.SerializeNumber(a.Amount)...)
	bytes = append(bytes, a.PaymentID...)

	wif := crypto.RIPEMD(bytes)[0:4]

	bytes = append(bytes, wif...)

	ret := base58.Encode(bytes)

	return prefix + ret, nil
}

//func DecodeAddr ( []byte ) (* Address, error){
//
//}

// tells whether address contains a paymentId
func (a *Address) IsIntegratedAddress() bool {
	return len(a.PaymentID) > 0
}

// if address has amount
func (a Address) IntegratedAmount() uint64 {
	return a.Amount
}
