package wizard

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction_simple_unstake"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"testing"
)

func TestCreateSimpleTx(t *testing.T) {

	dstPrivateKey := addresses.GenerateNewPrivateKey()
	dstAddress := dstPrivateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	dstAddressEncoded := dstAddress.EncodeAddr()

	privateKey := addresses.GenerateNewPrivateKey()
	tx := CreateSimpleTx(0, [][32]byte{privateKey.Key}, []uint64{1252}, [][]byte{{}}, []string{dstAddressEncoded}, []uint64{1250}, [][]byte{{}}, 0, []byte{})
	assert.NotNil(t, tx, "error creating simple tx")
	assert.NotPanics(t, tx.Validate, "error validating tx")
	assert.Equal(t, tx.VerifySignature(), true, "Verify signature failed")

	serialized := tx.Serialize()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NotPanics(t, func() { tx2.Deserialize(serialized) }, "deserialize failed")
	assert.NotPanics(t, tx2.Validate, "error validating tx")
	assert.Equal(t, tx2.VerifySignature(), true, "Verify signature failed2")

	fees := tx.ComputeFees()
	assert.Equal(t, fees[string([]byte{})], uint64(2), "Fees were calculated invalid")

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx := CreateUnstakeTx(0, privateKey.Key, 534, -1, []byte{}, false)
	assert.NotNil(t, tx, "creating unstake tx is nil")

	assert.NotPanics(t, tx.Validate, "error validating tx")

	assert.Equal(t, tx.VerifySignature(), true, "Verify signature failed")

	serialized := tx.Serialize()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NotPanics(t, func() { tx2.Deserialize(serialized) }, "deserialize failed")
	assert.NotPanics(t, tx2.Validate, "error validating tx")
	assert.Equal(t, tx2.VerifySignature(), true, "Verify signature failed2")

	fees := tx.ComputeFees()
	assert.Equal(t, fees[string(config.NATIVE_TOKEN)] > uint64(100), true, "Fees were calculated invalid")

	base := tx2.TxBase.(*transaction_simple.TransactionSimple)
	assert.Equal(t, fees[string(config.NATIVE_TOKEN)], base.Vin[0].Amount, "Fees are not paid by vin")

	unstake := base.Extra.(*transaction_simple_unstake.TransactionSimpleUnstake)
	assert.Equal(t, unstake.UnstakeAmount, uint64(534), "Fees are not paid by vin")
	assert.Equal(t, unstake.UnstakeFeeExtra, uint64(0), "Fees must be paid by vin")

}

func TestCreateUnstakeTxPayExtra(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx := CreateUnstakeTx(0, privateKey.Key, 534, -1, []byte{}, true)
	assert.NotNil(t, tx, "creating unstake tx is nil")

	assert.NotPanics(t, tx.Validate, "error validating tx")

	assert.Equal(t, tx.VerifySignature(), true, "Verify signature failed")

	serialized := tx.Serialize()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NotPanics(t, func() { tx2.Deserialize(serialized) }, "deserialize failed")
	assert.NotPanics(t, tx2.Validate, "error validating tx")
	assert.Equal(t, tx2.VerifySignature(), true, "Verify signature failed2")

	fees := tx.ComputeFees()
	assert.Equal(t, fees[string(config.NATIVE_TOKEN)] > uint64(100), true, "Fees were calculated invalid")

	base := tx2.TxBase.(*transaction_simple.TransactionSimple)
	assert.Equal(t, uint64(0), base.Vin[0].Amount, "Fees are not paid by vin")

	unstake := base.Extra.(*transaction_simple_unstake.TransactionSimpleUnstake)
	assert.Equal(t, unstake.UnstakeAmount, uint64(534), "Fees are not paid by vin")
	assert.Equal(t, unstake.UnstakeFeeExtra, fees[string(config.NATIVE_TOKEN)], "Fees are not paid by vin")
}
