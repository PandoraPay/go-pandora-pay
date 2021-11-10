package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

type zetherTxDataFrom struct {
	PrivateKey     helpers.HexBytes `json:"privateKey"`
	BalanceDecoded uint64           `json:"balanceDecoded"`
}

type zetherTxDataBase struct {
	From              []*zetherTxDataFrom                            `json:"from"`
	Assets            []helpers.HexBytes                             `json:"assets"`
	Amounts           []uint64                                       `json:"amounts"`
	Dsts              []string                                       `json:"dsts"`
	Burns             []uint64                                       `json:"burns"`
	RingMembers       [][]string                                     `json:"ringMembers"`
	Data              []*wizard.TransactionsWizardData               `json:"data"`
	Fees              []*wizard.TransactionsWizardFee                `json:"fees"`
	PayloadScriptType []transaction_zether_payload.PayloadScriptType `json:"payloadScriptType"`
	PayloadExtra      []wizard.WizardZetherPayloadExtra              `json:"payloadExtra"`
	Accs              map[string]map[string]helpers.HexBytes         `json:"accs"`
	Regs              map[string]helpers.HexBytes                    `json:"regs"`
	Height            uint64                                         `json:"height"`
	Hash              helpers.HexBytes                               `json:"hash"`
}

func prepareData(txData *zetherTxDataBase) (transfers []*wizard.WizardZetherTransfer, emap map[string]map[string][]byte, rings [][]*bn256.G1, publicKeyIndexes map[string]*wizard.WizardZetherPublicKeyIndex, err error) {

	assetsList := helpers.ConvertHexBytesArraysToBytesArray(txData.Assets)
	transfers = make([]*wizard.WizardZetherTransfer, len(txData.From))
	emap = wizard.InitializeEmap(assetsList)
	rings = make([][]*bn256.G1, len(txData.From))
	publicKeyIndexes = make(map[string]*wizard.WizardZetherPublicKeyIndex)

	for t, ast := range assetsList {

		key := addresses.PrivateKey{Key: txData.From[t].PrivateKey}

		var fromAddr *addresses.Address
		fromAddr, err = key.GenerateAddress(txData.Regs[string(key.GeneratePublicKey())] == nil, 0, nil)
		if err != nil {
			return
		}

		transfers[t] = &wizard.WizardZetherTransfer{
			Asset:              ast,
			From:               txData.From[t].PrivateKey,
			FromBalanceDecoded: txData.From[t].BalanceDecoded,
			Destination:        txData.Dsts[t],
			Amount:             txData.Amounts[t],
			Burn:               txData.Burns[t],
			Data:               txData.Data[t],
		}

		var payloadExtra wizard.WizardZetherPayloadExtra
		switch txData.PayloadScriptType[t] {
		case transaction_zether_payload.SCRIPT_TRANSFER:
			payloadExtra = nil
		case transaction_zether_payload.SCRIPT_CLAIM:
			payloadExtra = &wizard.WizardZetherPayloadExtraClaim{}
		case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
			payloadExtra = &wizard.WizardZetherPayloadExtraDelegateStake{}
		case transaction_zether_payload.SCRIPT_ASSET_CREATE:
			payloadExtra = &wizard.WizardZetherPayloadExtraAssetCreate{}
		case transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
			payloadExtra = &wizard.WizardZetherPayloadExtraAssetSupplyIncrease{}
		default:
			err = errors.New("Invalid PayloadScriptType")
			return
		}

		if payloadExtra != nil {
			var data []byte
			if data, err = json.Marshal(txData.PayloadExtra[t]); err != nil {
				return
			}
			if err = json.Unmarshal(data, payloadExtra); err != nil {
				return
			}
		}

		var x []byte
		x, err = json.Marshal(payloadExtra)
		fmt.Println(string(x))
		transfers[t].PayloadExtra = payloadExtra

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
				if acc, err = account.NewAccount(addr.PublicKey, ast); err != nil {
					return
				}
				if err = acc.Deserialize(helpers.NewBufferReader(accData)); err != nil {
					return
				}
				emap[string(ast)][p.G1().String()] = acc.Balance.Amount.Serialize()
			} else {
				var acckey crypto.Point
				if err = acckey.DecodeCompressed(addr.PublicKey); err != nil {
					return
				}
				emap[string(ast)][p.G1().String()] = crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G).Serialize()
			}

			ring = append(ring, p.G1())

			var reg *registration.Registration
			if regData := txData.Regs[hex.EncodeToString(addr.PublicKey)]; len(regData) > 0 {
				reg = registration.NewRegistration(addr.PublicKey)
				if err = reg.Deserialize(helpers.NewBufferReader(regData)); err != nil {
					return
				}
			}

			publicKeyIndex := &wizard.WizardZetherPublicKeyIndex{}
			publicKeyIndexes[string(addr.PublicKey)] = publicKeyIndex

			if reg != nil {
				publicKeyIndex.Registered = true
				publicKeyIndex.RegisteredIndex = reg.Index
			} else {
				if len(addr.Registration) == 0 {
					return fmt.Errorf("Signature is missing for %s", addr.EncodeAddr())
				}
				publicKeyIndex.RegistrationSignature = addr.Registration
			}

			return
		}

		if err = addPoint(fromAddr.EncodeAddr()); err != nil {
			return
		}
		if err = addPoint(txData.Dsts[t]); err != nil {
			return
		}
		for _, ringMember := range txData.RingMembers[t] {
			if err = addPoint(ringMember); err != nil {
				return
			}
		}

		rings[t] = ring
	}

	return
}

func createZetherTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 2 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a string and a callback")
		}

		txData := &zetherTxDataBase{}
		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		transfers, emap, rings, publicKeyIndexes, err := prepareData(txData)
		if err != nil {
			return nil, err
		}

		tx, err := wizard.CreateZetherTx(transfers, emap, rings, txData.Height, txData.Hash, publicKeyIndexes, txData.Fees, false, ctx, func(status string) {
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
