package wizard

import (
	"github.com/stretchr/testify/assert"
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
	tx, err := CreateSimpleTx(0, [][32]byte{privateKey.Key}, []uint64{1252}, [][]byte{{}}, []string{dstAddressEncoded}, []uint64{1250}, [][]byte{{}}, 0, []byte{})
	assert.NotNil(t, tx, "error creating simple tx")
	assert.Nil(t, err, "error creating simple tx")
	assert.Nil(t, tx.Validate(), "error validating tx")
	assert.Equal(t, tx.VerifySignature(), true, "Verify signature failed")

	serialized := tx.Serialize(true)
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.Nil(t, tx2.Deserialize(serialized), "deserialize failed")
	assert.Nil(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignature(), true, "Verify signature failed2")

	fees, err := tx.ComputeFees()
	assert.Nil(t, err, "Error validating fees")
	assert.Equal(t, fees[string([]byte{})], uint64(2), "Fees were calculated invalid")

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534, -1, []byte{})
	assert.NotNil(t, tx, "creating unstake tx is nil")
	assert.Nil(t, err, "error creating unstake tx")

	assert.Nil(t, tx.Validate(), "error validating tx")

	assert.Equal(t, tx.VerifySignature(), true, "Verify signature failed")

	serialized := tx.Serialize(true)
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.Nil(t, tx2.Deserialize(serialized), "deserialize failed")
	assert.Nil(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignature(), true, "Verify signature failed2")

	fees, err := tx.ComputeFees()
	assert.Nil(t, err, "Error validating fees")
	assert.Equal(t, fees[string([]byte{})] > uint64(100), true, "Fees were calculated invalid")

}
