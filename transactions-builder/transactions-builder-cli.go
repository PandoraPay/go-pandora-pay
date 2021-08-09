package transactions_builder

import (
	"bytes"
	"encoding/hex"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/transactions-builder/wizard"
)

func (builder *TransactionsBuilder) showWarningIfNotSyncCLI() {
	if builder.chain.Sync.GetSyncTime() == 0 {
		gui.GUI.OutputWrite("Your node is not Sync yet. Wait for it to get sync.")
	}
}

func (builder *TransactionsBuilder) readFees() (out *TransactionsBuilderFeeFloat, ok bool) {

	fee := &TransactionsBuilderFeeFloat{}

	if fee.PerByteAuto, ok = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte"); !ok {
		return
	}
	if !fee.PerByteAuto {
		if fee.PerByte, ok = gui.GUI.OutputReadFloat64("Fee per byte", nil); !ok {
			return
		}

		if fee.PerByte == 0 {
			if fee.Fixed, ok = gui.GUI.OutputReadFloat64("Fee per byte", nil); !ok {
				return
			}
		}
	}

	if fee.Fixed != 0 || fee.PerByte != 0 || fee.PerByteAuto {
		if fee.Token, ok = gui.GUI.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
			return
		}
	}

	return fee, true
}

func (builder *TransactionsBuilder) readFeesExtra() (out *TransactionsBuilderFeeFloatExtra, ok bool) {

	feeFloat, ok := builder.readFees()
	if !ok {
		return
	}

	fee := &TransactionsBuilderFeeFloatExtra{*feeFloat, false}
	if fee.PayInExtra, ok = gui.GUI.OutputReadBool("Pay in Extra. Type y/n"); !ok {
		return
	}

	return fee, true
}

func (builder *TransactionsBuilder) readData() (out *wizard.TransactionsWizardData, ok bool) {

	data := &wizard.TransactionsWizardData{}
	str, ok := gui.GUI.OutputReadString("Message (data). Leave empty for none")
	if !ok {
		return
	}

	if len(str) > 0 {
		data.Data = []byte(str)

		data.Encrypt, ok = gui.GUI.OutputReadBool("Encrypt message (data). Type y/n")

	}

	return data, ok
}

func (builder *TransactionsBuilder) initCLI() {

	cliTransfer := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer")
		if err != nil {
			return
		}

		token, ok := gui.GUI.OutputReadBytes("Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH})
		if !ok {
			return
		}
		if len(token) != 0 && len(token) != 40 {
			return errors.New("Invalid TokenId")
		}

		amount, ok := gui.GUI.OutputReadFloat64("Amount", nil)
		if !ok {
			return
		}

		destinationAddress, ok := gui.GUI.OutputReadAddress("Destination Address")
		if !ok {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		fee, ok := builder.readFees()
		if !ok {
			return
		}

		tx, err := builder.CreateSimpleTx_Float([]string{walletAddress.AddressEncoded}, nonce, []float64{amount}, [][]byte{token}, []string{destinationAddress.EncodeAddr()}, []float64{amount}, [][]byte{token}, data, fee, propagate, true, true, func(status string) {
			gui.GUI.OutputWrite(status)
		})

		// []float64{amount}, [][]byte{token}, feePerByte, feeToken, )
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		if propagate {
			gui.GUI.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	cliDelegate := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok := gui.GUI.OutputReadFloat64("Amount", nil)
		if !ok {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		delegateNewPublicKeyHashGenerate := false

		delegateNewPublicKeyHash, ok := gui.GUI.OutputReadBytes("Delegate New Public Key Hash. Use empty for not changing. Use '01' for generating a new one. ", []int{0, 1, cryptography.PublicKeyHashHashSize})
		if !ok {
			return
		}

		if len(delegateNewPublicKeyHash) == 1 {
			if bytes.Equal(delegateNewPublicKeyHash, []byte{1}) {
				delegateNewPublicKeyHash = []byte{}
				delegateNewPublicKeyHashGenerate = true
			} else {
				return errors.New("Invalid value for New Public key Hash")
			}
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		fee, ok := builder.readFees()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateDelegateTx_Float(walletAddress.AddressEncoded, nonce, amount, delegateNewPublicKeyHashGenerate, delegateNewPublicKeyHash, data, fee, propagate, true, true, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))
		if propagate {
			gui.GUI.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	cliUnstake := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok := gui.GUI.OutputReadFloat64("Amount", nil)
		if !ok {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave for automatically detection", nil, true)
		if !ok {
			return
		}

		data, ok := builder.readData()
		if !ok {
			return
		}

		fee, ok := builder.readFeesExtra()
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUnstakeTx_Float(walletAddress.AddressEncoded, nonce, amount, data, fee, propagate, true, true, func(status string) {
			gui.GUI.OutputWrite(status)
		})
		if err != nil {
			return
		}

		gui.GUI.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		if propagate {
			gui.GUI.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	gui.GUI.CommandDefineCallback("Transfer", cliTransfer, true)
	gui.GUI.CommandDefineCallback("Delegate", cliDelegate, true)
	gui.GUI.CommandDefineCallback("Unstake", cliUnstake, true)

}
