package transactions_builder

import (
	"encoding/hex"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
)

func (builder *TransactionsBuilder) initTransactionsBuilderCLI() {

	cliTransfer := func(cmd string) (err error) {

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

		amount, ok := gui.OutputReadFloat64("Amount")
		if !ok {
			return
		}

		feePerByte, ok := gui.OutputReadInt("Fee per byte. -1 automatically, 0 none")
		if !ok {
			return
		}

		var feeToken []byte
		if feePerByte != 0 {
			if feeToken, ok = gui.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
				return
			}
		}

		nonce, ok := gui.OutputReadUint64("Nonce. Leave 0 for automatically detection")
		if !ok {
			return
		}

		destinationAddress, ok := gui.OutputReadAddress("Destination Address")
		if !ok {
			return
		}

		tx, err := builder.CreateSimpleTx_Float([]string{walletAddress.AddressEncoded}, nonce, []float64{amount}, [][]byte{token}, []string{destinationAddress.EncodeAddr()}, []float64{amount}, [][]byte{token}, feePerByte, feeToken)
		if err != nil {
			return
		}

		gui.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		propagate, ok := gui.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		if propagate {
			result, err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height)
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

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Delegate")
		if err != nil {
			return
		}

		amount, ok := gui.OutputReadFloat64("Amount")
		if !ok {
			return
		}

		nonce, ok := gui.OutputReadUint64("Nonce. Leave 0 for automatically detection")
		if !ok {
			return
		}

		delegateNewPublicKeyHash, ok := gui.OutputReadBytes("Delegate New Public Key Hash. Leave it empty for not changing.", []int{0, cryptography.KeyHashSize})
		if !ok {
			return
		}

		feePerByte, ok := gui.OutputReadInt("Fee per byte. -1 automatically, 0 none")
		if !ok {
			return
		}

		var feeToken []byte
		if feePerByte != 0 {
			if feeToken, ok = gui.OutputReadBytes("Fee Token. Leave empty for Native Token", []int{0, config.TOKEN_LENGTH}); !ok {
				return
			}
		}

		tx, err := builder.CreateDelegateTx_Float(walletAddress.AddressEncoded, nonce, amount, delegateNewPublicKeyHash, feePerByte, feeToken)
		if err != nil {
			return
		}

		gui.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		propagate, ok := gui.OutputReadBool("Propagate. Type y/n")
		if !ok {
			return
		}

		if propagate {
			result, err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height)
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

	gui.CommandDefineCallback("Wallet : TX: Transfer", cliTransfer)
	gui.CommandDefineCallback("Wallet : TX: Delegate", cliDelegate)

}
