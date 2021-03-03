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
	tx, err := CreateSimpleTx(0, [][32]byte{privateKey.Key}, []uint64{1252}, [][]byte{{}}, []string{dstAddressEncoded}, []uint64{1250}, [][]byte{{}}, -1, []byte{})
	if err != nil {
		t.Errorf("error creating simple tx")
	}
	if err = tx.Validate(); err != nil {
		t.Errorf("error validating tx")
	}

	if tx.VerifySignature() == false {
		t.Errorf("Verify signature failed")
	}

	serialized := tx.Serialize(true)

	tx2 := new(transaction.Transaction)
	if err = tx2.Deserialize(serialized); err != nil {
		t.Errorf("Verify signature failed")
	}
	if err = tx2.Validate(); err != nil {
		t.Errorf("error validating tx")
	}

	if tx2.VerifySignature() == false {
		t.Errorf("Verify signature failed2")
	}

	fees, err := tx.ComputeFees()
	if err != nil {
		t.Errorf("Error validating fees")
	}
	if fees[string([]byte{})] != 2 {
		t.Errorf("Fees were calculated invalid")
	}

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534, -1, []byte{})
	if err != nil {
		t.Errorf("error creating unstake")
	}
	if err = tx.Validate(); err != nil {
		t.Errorf("error validating tx")
	}

	if tx.VerifySignature() == false {
		t.Errorf("Verify signature failed")
	}

	serialized := tx.Serialize(true)

	tx2 := new(transaction.Transaction)
	if err = tx2.Deserialize(serialized); err != nil {
		t.Errorf("Verify signature failed")
	}

	if err = tx2.Validate(); err != nil {
		t.Errorf("error validating tx")
	}

	if tx2.VerifySignature() == false {
		t.Errorf("Verify signature failed2")
	}

}
