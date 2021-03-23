package wizard

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"testing"
)

func TestCreateSimpleTx(t *testing.T) {

	dstPrivateKey := addresses.GenerateNewPrivateKey()
	dstAddress, err := dstPrivateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	assert.NoError(t, err)

	dstAddressEncoded := dstAddress.EncodeAddr()

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateSimpleTx(0, [][]byte{privateKey.Key}, []uint64{1252}, [][]byte{{}}, []string{dstAddressEncoded}, []uint64{1250}, [][]byte{{}}, 0, []byte{})
	assert.NoError(t, err)
	assert.NotNil(t, tx, "error creating simple tx")
	assert.NoError(t, tx.Validate(), "error validating tx")
	assert.Equal(t, tx.VerifySignatureManually(), true, "Verify signature failed")

	serialized := tx.Serialize()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)

	assert.NoError(t, tx2.Deserialize(helpers.NewBufferReader(serialized), true), "deserialize failed")
	assert.NoError(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignatureManually(), true, "Verify signature failed2")

	fees, err := tx.ComputeFees()
	assert.NoError(t, err)
	assert.Equal(t, fees[string([]byte{})], uint64(2), "Fees were calculated invalid")

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534, -1, []byte{}, false)
	assert.NoError(t, err)
	assert.NotNil(t, tx, "creating unstake tx is nil")

	assert.NoError(t, tx.Validate(), "error validating tx")

	assert.Equal(t, tx.VerifySignatureManually(), true, "Verify signature failed")

	serialized := tx.Serialize()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NoError(t, tx2.Deserialize(helpers.NewBufferReader(serialized), true), "deserialize failed")
	assert.NoError(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignatureManually(), true, "Verify signature failed2")

	fees, err := tx.ComputeFees()
	assert.NoError(t, err)
	assert.Equal(t, fees[config.NATIVE_TOKEN_STRING] > uint64(100), true, "Fees were calculated invalid")

	base := tx2.TxBase.(*transaction_simple.TransactionSimple)
	assert.Equal(t, fees[config.NATIVE_TOKEN_STRING], base.Vin[0].Amount, "Fees are not paid by vin")

	unstake := base.Extra.(*transaction_simple_extra.TransactionSimpleUnstake)
	assert.Equal(t, unstake.Amount, uint64(534), "Fees are not paid by vin")
	assert.Equal(t, unstake.FeeExtra, uint64(0), "Fees must be paid by vin")

}

func TestCreateUnstakeTxPayExtra(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534, -1, []byte{}, true)
	assert.NoError(t, err)
	assert.NotNil(t, tx, "creating unstake tx is nil")

	assert.NoError(t, tx.Validate(), "error validating tx")

	assert.Equal(t, tx.VerifySignatureManually(), true, "Verify signature failed")

	serialized := tx.Serialize()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NoError(t, tx2.Deserialize(helpers.NewBufferReader(serialized), true), "deserialize failed")
	assert.NoError(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignatureManually(), true, "Verify signature failed2")

	fees, err := tx.ComputeFees()
	assert.NoError(t, err)
	assert.Equal(t, fees[config.NATIVE_TOKEN_STRING] > uint64(100), true, "Fees were calculated invalid")

	base := tx2.TxBase.(*transaction_simple.TransactionSimple)
	assert.Equal(t, uint64(0), base.Vin[0].Amount, "Fees are not paid by vin")

	unstake := base.Extra.(*transaction_simple_extra.TransactionSimpleUnstake)
	assert.Equal(t, unstake.Amount, uint64(534), "Fees are not paid by vin")
	assert.Equal(t, unstake.FeeExtra, fees[config.NATIVE_TOKEN_STRING], "Fees are not paid by vin")
}
