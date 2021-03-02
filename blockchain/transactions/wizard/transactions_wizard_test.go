package wizard

import (
	"pandora-pay/addresses"
	"pandora-pay/helpers"
	"testing"
)

func TestCreateSimpleTx(t *testing.T) {

	dstPrivateKey := addresses.GenerateNewPrivateKey()
	dstAddress, _ := dstPrivateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	dstAddressEncoded, _ := dstAddress.EncodeAddr()

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateSimpleTx(0, [][32]byte{privateKey.Key}, []uint64{1252}, [][]byte{{0}}, []string{dstAddressEncoded}, []uint64{1252}, [][]byte{{0}})
	if err != nil {
		t.Errorf("error creating simple tx")
	}

	if tx.VerifySignature() == false {
		t.Errorf("Verify signature failed")
	}

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534)
	if err != nil {
		t.Errorf("error creating unstake")
	}

	if tx.VerifySignature() == false {
		t.Errorf("Verify signature failed")
	}

}
