package wizard

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/helpers"
	"testing"
)

func TestCreateUpdateDelegateTx(t *testing.T) {

	delegatedStakingPrivKey := addresses.GenerateNewPrivateKey()
	delegatedStakingUpdate := &transaction_data.TransactionDataDelegatedStakingUpdate{
		DelegatedStakingHasNewInfo:   true,
		DelegatedStakingNewPublicKey: delegatedStakingPrivKey.GeneratePublicKey(),
		DelegatedStakingNewFee:       10000,
	}
	delegatedStakingClaimAmount := uint64(0)

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateSimpleTx(0, privateKey.Key, 0, &WizardTxSimpleExtraUpdateDelegate{nil, delegatedStakingClaimAmount, delegatedStakingUpdate}, &TransactionsWizardData{[]byte{}, false}, &TransactionsWizardFee{PerByteAuto: true}, false, true, func(status string) {})
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

	fee, _ := tx.GetAllFee()
	assert.Equal(t, fee > uint64(100), true, "Fee were calculated invalid")

	base := tx2.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
	assert.Equal(t, fee, base.Fee, "Fee are not paid by vin")

	updateDelegate := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUpdateDelegate)
	assert.Equal(t, updateDelegate.DelegatedStakingUpdate.DelegatedStakingNewFee, delegatedStakingUpdate.DelegatedStakingNewFee, "Update delegate new fee is not set")
	assert.Equal(t, string(updateDelegate.DelegatedStakingUpdate.DelegatedStakingNewPublicKey), string(delegatedStakingUpdate.DelegatedStakingNewPublicKey), "Update delegate new public key is not set")

}

func TestCreateUnstakeTx(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	tx, err := CreateSimpleTx(0, privateKey.Key, 0, &WizardTxSimpleExtraUnstake{nil, 534}, &TransactionsWizardData{[]byte{}, false}, &TransactionsWizardFee{PerByteAuto: true}, false, true, func(status string) {})
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

	fee, _ := tx.GetAllFee()
	assert.Equal(t, fee > uint64(100), true, "Fee were calculated invalid")

	base := tx2.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
	assert.Equal(t, fee, base.Fee, "Fee are not paid by vin")

	unstake := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUnstake)
	assert.Equal(t, uint64(534), unstake.Amount, "Unstake amount is not set")

}
