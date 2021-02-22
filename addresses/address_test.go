package addresses

import (
	"pandora-pay/blockchain"
	"pandora-pay/helpers"
	"pandora-pay/helpers/base58"
	"testing"
)

func TestAddress_EncodeAddr(t *testing.T) {

	//WIF
	//1+20+1+4

	privateKey := GenerateNewPrivateKey()
	address, err := privateKey.GenerateTransparentAddress(true, 0, helpers.EmptyBytes(0))
	if err != nil || len(address.PublicKey) != 20 || len(address.PaymentID) != 0 {
		t.Errorf("Address Generation raised an error")
	}
	encoded, _ := address.EncodeAddr()
	decoded, err := base58.Decode(encoded[blockchain.NETWORK_BYTE_PREFIX_LENGTH:])
	if err != nil || len(decoded) != 1+20+1+4 {
		t.Errorf("AddressEncoded length is invalid")
	}

	address, err = privateKey.GenerateTransparentAddress(true, 20, helpers.EmptyBytes(0))
	if err != nil || len(address.PublicKey) != 20 || len(address.PaymentID) != 0 {
		t.Errorf("Address Generation raised an error")
	}

	encodedAmount, _ := address.EncodeAddr()
	if len(encoded) == len(encodedAmount) || encoded == encodedAmount {
		t.Errorf("Encoded Amounts are invalid")
	}

	address, err = privateKey.GenerateTransparentAddress(true, 20, helpers.EmptyBytes(8))
	if err != nil || len(address.PublicKey) != 20 || len(address.PaymentID) == 0 {
		t.Errorf("Address Generation raised an error")
	}

	encodedAmountPaymentId, _ := address.EncodeAddr()
	if len(encoded) == len(encodedAmount) || len(encodedAmount) == len(encodedAmountPaymentId) || encoded == encodedAmount || encodedAmount == encodedAmountPaymentId {
		t.Errorf("Encoded Amounts are invalid")
	}

}
