package wizard

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"testing"
)

func TestCreateSimpleTx(t *testing.T) {

	dstPrivateKey := addresses.GenerateNewPrivateKey()
	dstAddress, _ := dstPrivateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	dstAddressEncoded, _ := dstAddress.EncodeAddr()

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateSimpleTx(0, [][32]byte{privateKey.Key}, []uint64{1252}, [][]byte{{}}, []string{dstAddressEncoded}, []uint64{1252}, [][]byte{{}}, 0, 1)
	if err != nil {
		t.Errorf("error creating simple tx")
	}

	if tx.VerifySignature() == false {
		t.Errorf("Verify signature failed")
	}

	serialized := tx.Serialize(true)

	tx2 := new(transaction.Transaction)
	if err = tx2.Deserialize(serialized); err != nil {
		t.Errorf("Verify signature failed")
	}

	if tx2.VerifySignature() == false {
		t.Errorf("Verify signature failed2")
	}

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534, 0, 1)
	if err != nil {
		t.Errorf("error creating unstake")
	}

	if tx.VerifySignature() == false {
		t.Errorf("Verify signature failed")
	}

	serialized := tx.Serialize(true)

	tx2 := new(transaction.Transaction)
	if err = tx2.Deserialize(serialized); err != nil {
		t.Errorf("Verify signature failed")
	}

	if tx2.VerifySignature() == false {
		t.Errorf("Verify signature failed2")
	}

}
