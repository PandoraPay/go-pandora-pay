package wizard

import (
	"github.com/stretchr/testify/assert"
	"pandora-pay/addresses"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"testing"
)

func TestWizardSimple_CreateTx(t *testing.T) {

	vin1 := addresses.GenerateNewPrivateKey()
	vin2 := addresses.GenerateNewPrivateKey()

	vout1 := addresses.GenerateNewPrivateKey()
	vout2 := addresses.GenerateNewPrivateKey()
	vout3 := addresses.GenerateNewPrivateKey()

	asset1 := helpers.RandomBytes(config_coins.ASSET_LENGTH)
	asset2 := helpers.RandomBytes(config_coins.ASSET_LENGTH)

	tx, err := CreateSimpleTx(&WizardTxSimpleTransfer{
		nil,
		&WizardTransactionData{Data: helpers.RandomBytes(20)},
		&WizardTransactionFee{},
		50,
		[]*WizardTxSimpleTransferVin{
			{
				vin1.Key,
				30,
				asset1,
			},
			{
				vin2.Key,
				20,
				asset2,
			},
		},
		[]*WizardTxSimpleTransferVout{
			{
				vout1.GeneratePublicKeyHash(),
				5,
				asset1,
			},
			{
				vout2.GeneratePublicKeyHash(),
				20,
				asset2,
			},
			{
				vout3.GeneratePublicKeyHash(),
				3,
				asset1,
			},
		},
	}, true, func(string) {})
	assert.Nil(t, err)

	assert.NotNil(t, tx)

	assert.Nil(t, tx.Validate())

	assert.Equal(t, tx.VerifySignatureManually(), true)

}
