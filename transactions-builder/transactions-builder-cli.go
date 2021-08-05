package transactions_builder

import (
	"bytes"
	"encoding/hex"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
)

func (builder *TransactionsBuilder) showWarningIfNotSyncCLI() {
	if builder.chain.Sync.GetSyncTime() == 0 {
		gui.GUI.OutputWrite("Your node is not Sync yet. Wait for it to get sync.")
	}
}

func (builder *TransactionsBuilder) readFees(allowFeesPayInExtra bool) (feeFixed, feePerByte float64, feePerByteAuto bool, feeToken []byte, feePayInExtra, ok bool) {

	if feePerByteAuto, ok = gui.GUI.OutputReadBool("Compute Automatically Fee Per Byte"); !ok {
		return
	}
	if !feePerByteAuto {
		if feePerByte, ok = gui.GUI.OutputReadFloat64("Fee per byte", nil); !ok {
			return
		}

		if feePerByte == 0 {
			if feeFixed, ok = gui.GUI.OutputReadFloat64("Fee per byte", nil); !ok {
				return
			}
		}
	}

	if feeFixed != 0 || feePerByte != 0 || feePerByteAuto {
		if feeToken, ok = gui.GUI.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
			return
		}
	}

	if allowFeesPayInExtra {
		feePayInExtra, ok = gui.GUI.OutputReadBool("Pay in Extra. Type y/n")
		if !ok {
			return
		}
	}

	return
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

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		feeFixed, feePerByte, feePerByteAuto, feeToken, _, ok := builder.readFees(false)
		if !ok {
			return
		}

		tx, err := builder.CreateSimpleTx_Float([]string{walletAddress.AddressEncoded}, nonce, []float64{amount}, [][]byte{token}, []string{destinationAddress.EncodeAddr()}, []float64{amount}, [][]byte{token}, feeFixed, feePerByte, feePerByteAuto, feeToken, propagate, true, true, func(status string) {
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

		feeFixed, feePerByte, feePerByteAuto, feeToken, _, ok := builder.readFees(false)
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateDelegateTx_Float(walletAddress.AddressEncoded, nonce, amount, delegateNewPublicKeyHashGenerate, delegateNewPublicKeyHash, feeFixed, feePerByte, feePerByteAuto, feeToken, propagate, true, true, func(status string) {
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

		feeFixed, feePerByte, feePerByteAuto, feeToken, feePayInExtra, ok := builder.readFees(true)
		if !ok {
			return
		}

		propagate, ok := gui.GUI.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUnstakeTx_Float(walletAddress.AddressEncoded, nonce, amount, feeFixed, feePerByte, feePerByteAuto, feeToken, feePayInExtra, propagate, true, true, func(status string) {
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
