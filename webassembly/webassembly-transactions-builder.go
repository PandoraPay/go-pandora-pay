package webassembly

import (
	"context"
	"encoding/json"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/app"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common/api_types"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/transactions-builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
	"time"
)

func createZetherTx_Float(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		txData := &struct {
			From        []string                                            `json:"from"`
			Tokens      []helpers.HexBytes                                  `json:"tokens"`
			Amounts     []float64                                           `json:"amounts"`
			Dsts        []string                                            `json:"dsts"`
			Burns       []float64                                           `json:"burns"`
			RingMembers [][]string                                          `json:"ringMembers"`
			Data        []*wizard.TransactionsWizardData                    `json:"data"`
			Fees        []*transactions_builder.TransactionsBuilderFeeFloat `json:"fees"`
			PropagateTx bool                                                `json:"propagateTx"`
			AwaitAnswer bool                                                `json:"awaitAnswer"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		socket := app.Network.Websockets.GetFirstSocket()
		if socket == nil {
			return nil, errors.New("You are not connected to any node")
		}

		emap := make(map[string]map[string]*account.Account)
		regsMap := make(map[string]*registration.Registration)

		for t, ring := range txData.RingMembers {

			if emap[string(txData.Tokens[t])] == nil {
				emap[string(txData.Tokens[t])] = make(map[string]*account.Account)
			}

			shuffle := helpers.ShuffleArray_for_Zether(len(ring))

			request := &api_types.APIAccountsByKeysRequest{
				Keys:       make([]*api_types.APIAccountBaseRequest, len(ring)),
				ReturnType: api_types.RETURN_SERIALIZED,
				Token:      txData.Tokens[t],
			}

			for i := 0; i < len(shuffle); i++ {
				addr, err := addresses.DecodeAddr(ring[shuffle[i]])
				if err != nil {
					return nil, err
				}
				request.Keys[i] = &api_types.APIAccountBaseRequest{
					PublicKey: addr.PublicKey,
				}
			}

			data := socket.SendJSONAwaitAnswer([]byte("accounts/by-keys"), request, 0)
			if data.Err != nil {
				return nil, data.Err
			}

			answer := &api_types.APIAccountsByKeys{}
			if err := json.Unmarshal(data.Out, answer); err != nil {
				return nil, err
			}

			for i, key := range request.Keys {
				var acc *account.Account
				if len(answer.AccSerialized[i]) > 0 {
					acc = account.NewAccount(key.PublicKey, txData.Tokens[t])
					if err := acc.Deserialize(helpers.NewBufferReader(answer.AccSerialized[i])); err != nil {
						return nil, err
					}
				}
				emap[string(txData.Tokens[t])][string(key.PublicKey)] = acc

				var reg *registration.Registration
				if len(answer.RegSerialized[i]) > 0 {
					reg = registration.NewRegistration(key.PublicKey)
					if err := reg.Deserialize(helpers.NewBufferReader(answer.RegSerialized[i])); err != nil {
						return nil, err
					}
				}
				regsMap[string(key.PublicKey)] = reg
			}

		}

		tx, err := app.TransactionsBuilder.CreateZetherTx_Float(txData.From, helpers.ConvertHexBytesArraysToBytesArray(txData.Tokens), txData.Amounts, txData.Dsts, txData.Burns, txData.RingMembers, txData.Data, txData.Fees, txData.PropagateTx, txData.AwaitAnswer, false, func(dataStorage *data_storage.DataStorage) error {

			for token, value := range emap {
				accs, err := dataStorage.AccsCollection.GetMap([]byte(token))
				if err != nil {
					return err
				}

				for publicKey, acc := range value {
					if err := accs.UpdateOrDelete(publicKey, acc); err != nil {
						return err
					}
				}
			}

			for publicKey, reg := range regsMap {
				if err := dataStorage.Regs.UpdateOrDelete(publicKey, reg); err != nil {
					return err
				}
			}

			return nil
		}, ctx, func(status string) {
			args[1].Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)
	})
}

func createUpdateDelegateTx_Float(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		type DelegateTxFloatData struct {
			From                         string                                            `json:"from"`
			Nonce                        uint64                                            `json:"nonce"`
			DelegateNewPublicKeyGenerate bool                                              `json:"delegateNewPublicKeyGenerate"`
			DelegateNewPubKey            helpers.HexBytes                                  `json:"delegateNewPubKey"`
			DelegateNewFee               uint64                                            `json:"delegateNewFee"`
			Data                         *wizard.TransactionsWizardData                    `json:"data"`
			Fee                          *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx                  bool                                              `json:"propagateTx"`
			AwaitAnswer                  bool                                              `json:"awaitAnswer"`
		}

		txData := &DelegateTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUpdateDelegateTx_Float(txData.From, txData.Nonce, txData.DelegateNewPublicKeyGenerate, txData.DelegateNewPubKey, txData.DelegateNewFee, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			args[1].Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}

func createUnstakeTx_Float(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		type DelegateTxFloatData struct {
			From          string                                            `json:"from"`
			Nonce         uint64                                            `json:"nonce"`
			UnstakeAmount float64                                           `json:"unstakeAmount"`
			Data          *wizard.TransactionsWizardData                    `json:"data"`
			Fee           *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx   bool                                              `json:"propagateTx"`
			AwaitAnswer   bool                                              `json:"awaitAnswer"`
		}

		txData := &DelegateTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUnstakeTx_Float(txData.From, txData.Nonce, txData.UnstakeAmount, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			args[1].Invoke(status)
			time.Sleep(1 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}
