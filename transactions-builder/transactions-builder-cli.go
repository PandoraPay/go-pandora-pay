package transactions_builder

import (
	"encoding/hex"
	"errors"
	"pandora-pay/gui"
)

func (builder *TransactionsBuilder) initTransactionsBuilderCLI() {

	cliTransfer := func(cmd string) (err error) {

		walletAddress, _, err := builder.wallet.CliSelectAddress("Select Address to Transfer")
		if err != nil {
			return
		}

		amount, ok := gui.OutputReadUint64("Amount")
		if !ok {
			return
		}

		feePerByte, ok := gui.OutputReadInt("Fee per byte. -1 automatically, 0 none")
		if !ok {
			return
		}

		token, ok := gui.OutputReadToken("Token. Leave empty for the Native Token")
		if !ok {
			return
		}
		if len(token) != 0 && len(token) != 40 {
			return errors.New("Invalid TokenId")
		}

		nonce, ok := gui.OutputReadUint64("Nonce. Leave 0 for automatically detection")
		if !ok {
			return
		}

		destinationAddress, ok := gui.OutputReadAddress("Destination Address")
		if !ok {
			return
		}

		tx, err := builder.CreateSimpleTx([]string{walletAddress.AddressEncoded}, nonce, []uint64{amount}, [][]byte{token}, []string{destinationAddress.EncodeAddr()}, []uint64{amount}, [][]byte{token}, feePerByte, token)
		if err != nil {
			return
		}

		gui.OutputWrite("Tx created: " + hex.EncodeToString(tx.Bloom.Hash))

		propagate, ok := gui.OutputReadBool("Destination Address")
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
		}

		return
	}

	gui.CommandDefineCallback("Wallet : TX: Transfer", cliTransfer)

}
