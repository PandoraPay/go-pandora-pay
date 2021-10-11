package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
	"time"
)

func createZetherTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 2 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a string and a callback")
		}

		txData := &struct {
			FromPrivateKeys     []helpers.HexBytes                     `json:"fromPrivateKeys"`
			FromBalancesDecoded []uint64                               `json:"fromBalancesDecoded"`
			Assets              []helpers.HexBytes                     `json:"assets"`
			Amounts             []uint64                               `json:"amounts"`
			Dsts                []string                               `json:"dsts"`
			Burns               []uint64                               `json:"burns"`
			RingMembers         [][]string                             `json:"ringMembers"`
			Data                []*wizard.TransactionsWizardData       `json:"data"`
			Fees                []*wizard.TransactionsWizardFee        `json:"fees"`
			Accs                map[string]map[string]helpers.HexBytes `json:"accs"`
			Regs                map[string]helpers.HexBytes            `json:"regs"`
			Height              uint64                                 `json:"height"`
			Hash                helpers.HexBytes                       `json:"hash"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		assetsList := helpers.ConvertHexBytesArraysToBytesArray(txData.Assets)

		transfers := make([]*wizard.ZetherTransfer, len(txData.FromPrivateKeys))
		emap := wizard.InitializeEmap(assetsList)
		rings := make([][]*bn256.G1, len(txData.FromPrivateKeys))
		publicKeyIndexes := make(map[string]*wizard.ZetherPublicKeyIndex)

		for t, ast := range assetsList {

			key := addresses.PrivateKey{Key: txData.FromPrivateKeys[t]}
			fromAddr, err := key.GenerateAddress(false, 0, nil)
			if err != nil {
				return nil, err
			}

			transfers[t] = &wizard.ZetherTransfer{
				Asset:              ast,
				From:               txData.FromPrivateKeys[t],
				FromBalanceDecoded: txData.FromBalancesDecoded[t],
				Destination:        txData.Dsts[t],
				Amount:             txData.Amounts[t],
				Burn:               txData.Burns[t],
				Data:               txData.Data[t],
			}

			uniqueMap := make(map[string]bool)
			var ring []*bn256.G1

			addPoint := func(address string) (err error) {

				var addr *addresses.Address
				var p *crypto.Point

				if addr, err = addresses.DecodeAddr(address); err != nil {
					return
				}
				if uniqueMap[string(addr.PublicKey)] {
					return
				}
				uniqueMap[string(addr.PublicKey)] = true

				if p, err = addr.GetPoint(); err != nil {
					return
				}

				var acc *account.Account
				if accData := txData.Accs[hex.EncodeToString(ast)][hex.EncodeToString(addr.PublicKey)]; len(accData) > 0 {
					acc = account.NewAccount(addr.PublicKey, ast)
					if err = acc.Deserialize(helpers.NewBufferReader(accData)); err != nil {
						return
					}
					emap[string(ast)][p.G1().String()] = acc.Balance.Amount.Serialize()
				} else {
					emap[string(ast)][p.G1().String()] = crypto.ConstructElGamal(p.G1(), crypto.ElGamal_BASE_G).Serialize()
				}

				ring = append(ring, p.G1())

				var reg *registration.Registration
				if regData := txData.Regs[hex.EncodeToString(addr.PublicKey)]; len(regData) > 0 {
					reg = registration.NewRegistration(addr.PublicKey)
					if err = reg.Deserialize(helpers.NewBufferReader(regData)); err != nil {
						return
					}
				}

				publicKeyIndex := &wizard.ZetherPublicKeyIndex{}
				publicKeyIndexes[string(addr.PublicKey)] = publicKeyIndex

				if reg != nil {
					publicKeyIndex.Registered = true
					publicKeyIndex.RegisteredIndex = reg.Index
				} else {
					publicKeyIndex.RegistrationSignature = addr.Registration
				}

				return
			}

			if err = addPoint(fromAddr.EncodeAddr()); err != nil {
				return nil, err
			}
			if err = addPoint(txData.Dsts[t]); err != nil {
				return nil, err
			}
			for _, ringMember := range txData.RingMembers[t] {
				if err = addPoint(ringMember); err != nil {
					return nil, err
				}
			}

			rings[t] = ring
		}

		tx, err := wizard.CreateZetherTx(transfers, emap, rings, txData.Height, txData.Hash, publicKeyIndexes, txData.Fees, false, ctx, func(status string) {
			args[1].Invoke(status)
			time.Sleep(10 * time.Millisecond)
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
