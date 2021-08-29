package wizard

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/helpers"
	"testing"
)

func TestCreateUpdateDelegateTx(t *testing.T) {

	delegatePrivKey := addresses.GenerateNewPrivateKey()
	delegatePubKey := delegatePrivKey.GeneratePublicKey()
	delegateFee := uint64(10000)

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUpdateDelegateTx(0, privateKey.Key, delegatePubKey, delegateFee, &TransactionsWizardData{[]byte{}, false}, &TransactionsWizardFee{PerByteAuto: true}, func(status string) {})
	assert.NoError(t, err)
	assert.NotNil(t, tx, "creating update delegate tx is nil")

	assert.NoError(t, tx.Validate(), "error validating tx")

	assert.Equal(t, tx.VerifySignatureManually(), true, "Verify signature failed")

	serialized := tx.SerializeManualToBytes()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NoError(t, tx2.Deserialize(helpers.NewBufferReader(serialized)), "deserialize failed")
	assert.NoError(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignatureManually(), true, "Verify signature failed2")

	fees := tx.GetAllFees()
	assert.Equal(t, fees > uint64(100), true, "Fees were calculated invalid")

	base := tx2.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
	assert.Equal(t, fees, base.Fee, "Fees are not paid by vin")

	updateDelegate := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUpdateDelegate)
	assert.Equal(t, updateDelegate.NewFee, delegateFee, "Update delegate new fee is not set")
	assert.Equal(t, string(updateDelegate.NewPublicKey), string(delegatePubKey), "Update delegate new public key is not set")

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateUnstakeTx(0, privateKey.Key, 534, &TransactionsWizardData{[]byte{}, false}, &TransactionsWizardFee{PerByteAuto: true}, func(status string) {})
	assert.NoError(t, err)
	assert.NotNil(t, tx, "creating unstake tx is nil")

	assert.NoError(t, tx.Validate(), "error validating tx")

	assert.Equal(t, tx.VerifySignatureManually(), true, "Verify signature failed")

	serialized := tx.SerializeManualToBytes()
	assert.NotNil(t, serialized, "serialized is nil")

	tx2 := new(transaction.Transaction)
	assert.NoError(t, tx2.Deserialize(helpers.NewBufferReader(serialized)), "deserialize failed")
	assert.NoError(t, tx2.Validate(), "error validating tx")
	assert.Equal(t, tx2.VerifySignatureManually(), true, "Verify signature failed2")

	fees := tx.GetAllFees()
	assert.Equal(t, fees > uint64(100), true, "Fees were calculated invalid")

	base := tx2.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
	assert.Equal(t, fees, base.Fee, "Fees are not paid by vin")

	unstake := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake)
	assert.Equal(t, uint64(534), unstake.Amount, "Unstake amount is not set")

}
