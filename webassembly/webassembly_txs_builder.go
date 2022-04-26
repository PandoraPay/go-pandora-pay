package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/txs_builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

func createSimpleTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		txData := &struct {
			TxScript transaction_simple.ScriptType `json:"txScript"`
			Sender   string                        `json:"sender"`
			Nonce    uint64                        `json:"nonce"`
			Extra    wizard.WizardTxSimpleExtra    `json:"extra"`
			Data     *wizard.WizardTransactionData `json:"data"`
			Fee      *wizard.WizardTransactionFee  `json:"fee"`
			Height   uint64                        `json:"height"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		var payloadExtra wizard.WizardTxSimpleExtra
		switch txData.TxScript {
		case transaction_simple.SCRIPT_TRANSFER:
		default:
			return nil, errors.New("Invalid PayloadScriptType")
		}

		if payloadExtra != nil {
			data, err := json.Marshal(txData.Extra)
			if err != nil {
				return nil, err
			}
			if err = json.Unmarshal(data, payloadExtra); err != nil {
				return nil, err
			}
		}

		senderWalletAddr, err := app.Wallet.GetWalletAddressByEncodedAddress(txData.Sender, true)
		if err != nil {
			return nil, err
		}

		if senderWalletAddr.PrivateKey.Key == nil {
			return nil, errors.New("Can't be used for transactions as the private key is missing")
		}

		tx, err := wizard.CreateSimpleTx(&wizard.WizardTxSimpleTransfer{
			payloadExtra,
			txData.Data,
			txData.Fee,
			txData.Nonce,
			senderWalletAddr.PrivateKey.Key,
		}, false, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		txJson, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}

		return []interface{}{
			webassembly_utils.ConvertBytes(txJson),
			webassembly_utils.ConvertBytes(tx.Bloom.Serialized),
		}, nil

	})
}
