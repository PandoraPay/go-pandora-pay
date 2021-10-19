package webassembly

import (
	"errors"
	"pandora-pay/app"
	"pandora-pay/helpers"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

func createUpdateDelegateTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		txData := &struct {
			From                         string                         `json:"from"`
			Nonce                        uint64                         `json:"nonce"`
			DelegatedStakingNewPublicKey helpers.HexBytes               `json:"delegatedStakingNewPublicKey"`
			DelegatedStakingNewFee       uint64                         `json:"delegatedStakingNewFee"`
			DelegatedStakingClaimAmount  uint64                         `json:"delegatedStakingClaimAmount"`
			Data                         *wizard.TransactionsWizardData `json:"data"`
			Fee                          *wizard.TransactionsWizardFee  `json:"fee"`
			PropagateTx                  bool                           `json:"propagateTx"`
			AwaitAnswer                  bool                           `json:"awaitAnswer"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUpdateDelegateTx(txData.From, txData.Nonce, txData.DelegatedStakingNewPublicKey, txData.DelegatedStakingNewFee, txData.DelegatedStakingClaimAmount, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, false, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}

func createUnstakeTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		txData := &struct {
			From          string                         `json:"from"`
			Nonce         uint64                         `json:"nonce"`
			UnstakeAmount uint64                         `json:"unstakeAmount"`
			Data          *wizard.TransactionsWizardData `json:"data"`
			Fee           *wizard.TransactionsWizardFee  `json:"fee"`
			PropagateTx   bool                           `json:"propagateTx"`
			AwaitAnswer   bool                           `json:"awaitAnswer"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUnstakeTx(txData.From, txData.Nonce, txData.UnstakeAmount, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, false, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}
