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
		gui.OutputWrite("Your node is not Sync yet. Wait for it to get sync.")
	}
}

func (builder *TransactionsBuilder) initCLI() {

	cliTransfer := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer")
		if err != nil {
			return
		}

		token, ok := gui.OutputReadBytes("Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH})
		if !ok {
			return
		}
		if len(token) != 0 && len(token) != 40 {
			return errors.New("Invalid TokenId")
		}

		amount, ok := gui.OutputReadFloat64("Amount", nil)
		if !ok {
			return
		}

		feePerByte, ok := gui.OutputReadInt("Fee per byte. -1 automatically, 0 none", nil)
		if !ok {
			return
		}

		var feeToken []byte
		if feePerByte != 0 {
			if feeToken, ok = gui.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
				return
			}
		}

		nonce, ok := gui.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		destinationAddress, ok := gui.OutputReadAddress("Destination Address")
		if !ok {
			return
		}

		tx, err := builder.CreateSimpleTx_Float([]string{walletAddress.GetAddressEncoded()}, nonce, []float64{amount}, [][]byte{token}, []string{destinationAddress.EncodeAddr()}, []float64{amount}, [][]byte{token}, feePerByte, feeToken)
		if err != nil {
			return
		}

		gui.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		propagate, ok := gui.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		if propagate {
			result, err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, true)
			if err != nil {
				return err
			}
			if !result {
				return errors.New("transaction was not inserted in mempool")
			}
			gui.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	cliDelegate := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok := gui.OutputReadFloat64("Amount", nil)
		if !ok {
			return
		}

		nonce, ok := gui.OutputReadUint64("Nonce. Leave empty for automatically detection", nil, true)
		if !ok {
			return
		}

		delegateNewPublicKeyHashGenerate := false

		delegateNewPublicKeyHash, ok := gui.OutputReadBytes("Delegate New Public Key Hash. Use empty for not changing. Use '01' for generating a new one. ", []int{0, 1, cryptography.KeyHashSize})
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

		feePerByte, ok := gui.OutputReadInt("Fee per byte. -1 automatically, 0 none", nil)
		if !ok {
			return
		}

		var feeToken []byte
		if feePerByte != 0 {
			if feeToken, ok = gui.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
				return
			}
		}

		tx, err := builder.CreateDelegateTx_Float(walletAddress.GetAddressEncoded(), nonce, amount, delegateNewPublicKeyHashGenerate, delegateNewPublicKeyHash, feePerByte, feeToken)
		if err != nil {
			return
		}

		gui.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		propagate, ok := gui.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		if propagate {
			result, err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, true)
			if err != nil {
				return err
			}
			if !result {
				return errors.New("transaction was not inserted in mempool")
			}
			gui.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	cliUnstake := func(cmd string) (err error) {

		builder.showWarningIfNotSyncCLI()

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok := gui.OutputReadFloat64("Amount", nil)
		if !ok {
			return
		}

		nonce, ok := gui.OutputReadUint64("Nonce. Leave for automatically detection", nil, true)
		if !ok {
			return
		}

		feePerByte, ok := gui.OutputReadInt("Fee per byte. -1 automatically, 0 none", nil)
		if !ok {
			return
		}

		var feeToken []byte
		if feePerByte != 0 {
			if feeToken, ok = gui.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
				return
			}
		}

		payFeeInExtra, ok := gui.OutputReadBool("Pay in Extra. Type y/n")
		if !ok {
			return
		}

		tx, err := builder.CreateUnstakeTx_Float(walletAddress.GetAddressEncoded(), nonce, amount, feePerByte, feeToken, payFeeInExtra)
		if err != nil {
			return
		}

		gui.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		propagate, ok := gui.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		if propagate {
			result, err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, true)
			if err != nil {
				return err
			}
			if !result {
				return errors.New("transaction was not inserted in mempool")
			}
			gui.OutputWrite("Tx was inserted in mempool")
		}

		return
	}

	gui.CommandDefineCallback("Transfer", cliTransfer)
	gui.CommandDefineCallback("Delegate", cliDelegate)
	gui.CommandDefineCallback("Unstake", cliUnstake)

}
