package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/transactions_builder/wizard"
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
			TxScript                    transaction_simple.ScriptType                           `json:"txScript"`
			From                        string                                                  `json:"from"`
			Nonce                       uint64                                                  `json:"nonce"`
			DelegatedStakingClaimAmount uint64                                                  `json:"delegatedStakingClaimAmount"`
			DelegatedStakingUpdate      *transaction_data.TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
			Extra                       wizard.TxTransferSimpleExtra                            `json:"extra"`
			Data                        *wizard.TransactionsWizardData                          `json:"data"`
			Fee                         *wizard.TransactionsWizardFee                           `json:"fee"`
			PropagateTx                 bool                                                    `json:"propagateTx"`
			AwaitAnswer                 bool                                                    `json:"awaitAnswer"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		var payloadExtra wizard.TxTransferSimpleExtra
		switch txData.TxScript {
		case transaction_simple.SCRIPT_UNSTAKE:
			payloadExtra = &wizard.TxTransferSimpleExtraUnstake{}
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:
			payloadExtra = &wizard.TxTransferSimpleExtraUpdateDelegate{}
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

		tx, err := app.TransactionsBuilder.CreateSimpleTx(txData.From, txData.Nonce, payloadExtra, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, false, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}
