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

	if fee.PerByteAuto, ok = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte. Type y\n"); !ok {
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

		ringMembers := make([][]string, 1)
		if ringMembers[0], err = builder.CreateZetherRing(walletAddress.AddressEncoded, destinationAddress.EncodeAddr(), token, -1, -1); err != nil {
			return
		}

		tx, err := builder.CreateZetherTx_Float([]string{walletAddress.AddressEncoded}, [][]byte{token}, []float64{amount}, []string{destinationAddress.EncodeAddr()}, []float64{0}, ringMembers, []*wizard.TransactionsWizardData{data}, []*TransactionsBuilderFeeFloat{fee}, propagate, true, true, nil, func(status string) {
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

	cliDelegate := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		nonce, ok := gui.GUI.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		delegateNewPublicKeyGenerate := false

		delegateNewPublicKey, ok := gui.GUI.OutputReadBytes("Delegate New Public Key. Use empty for not changing. Use '01' for generating a new one. ", []int{0, 1, cryptography.PublicKeySize})
		if !ok {
			return
		}

		if len(delegateNewPublicKey) == 1 {
			if bytes.Equal(delegateNewPublicKey, []byte{1}) {
				delegateNewPublicKey = []byte{}
				delegateNewPublicKeyGenerate = true
			} else {
				return errors.New("Invalid value for New Public key Hash")
			}
		}

		var delegateNewFee uint64
		if len(delegateNewPublicKey) > 0 {
			number, ok := gui.GUI.OutputReadUint64("New Fee", nil, true)
			if !ok {
				return
			}
			delegateNewFee = number
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

		tx, err := builder.CreateUpdateDelegateTx_Float(walletAddress.AddressEncoded, nonce, delegateNewPublicKeyGenerate, delegateNewPublicKey, delegateNewFee, data, fee, propagate, true, true, func(status string) {
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

		fee, ok := builder.readFees()
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
