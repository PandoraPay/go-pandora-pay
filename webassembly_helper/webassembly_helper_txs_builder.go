package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/txs_builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

type zetherTxDataSender struct {
	PrivateKey       []byte `json:"privateKey"`
	DecryptedBalance uint64 `json:"decryptedBalance"`
}

type zetherTxDataBase struct {
	Senders           []*zetherTxDataSender                          `json:"senders"`
	Assets            [][]byte                                       `json:"assets"`
	Amounts           []uint64                                       `json:"amounts"`
	Recipients        []string                                       `json:"recipients"`
	Burns             []uint64                                       `json:"burns"`
	RingMembers       [][]string                                     `json:"ringMembers"`
	Data              []*wizard.WizardTransactionData                `json:"data"`
	Fees              []*wizard.WizardZetherTransactionFee           `json:"fees"`
	PayloadScriptType []transaction_zether_payload.PayloadScriptType `json:"payloadScriptType"`
	PayloadExtra      []wizard.WizardZetherPayloadExtra              `json:"payloadExtra"`
	Accs              map[string]map[string][]byte                   `json:"accs"`
	Regs              map[string][]byte                              `json:"regs"`
	Height            uint64                                         `json:"height"`
	Hash              []byte                                         `json:"hash"`
}

func prepareData(txData *zetherTxDataBase) (transfers []*wizard.WizardZetherTransfer, emap map[string]map[string][]byte, rings [][]*bn256.G1, publicKeyIndexes map[string]*wizard.WizardZetherPublicKeyIndex, err error) {

	transfers = make([]*wizard.WizardZetherTransfer, len(txData.Senders))
	emap = wizard.InitializeEmap(txData.Assets)
	rings = make([][]*bn256.G1, len(txData.Senders))
	publicKeyIndexes = make(map[string]*wizard.WizardZetherPublicKeyIndex)

	for t, ast := range txData.Assets {

		key := addresses.PrivateKey{Key: txData.Senders[t].PrivateKey}

		var senderAddr *addresses.Address
		senderAddr, err = key.GenerateAddress(txData.Regs[string(key.GeneratePublicKey())] == nil, nil, 0, nil)
		if err != nil {
			return
		}

		transfers[t] = &wizard.WizardZetherTransfer{
			Asset:                  ast,
			Sender:                 txData.Senders[t].PrivateKey,
			SenderDecryptedBalance: txData.Senders[t].DecryptedBalance,
			Recipient:              txData.Recipients[t],
			Amount:                 txData.Amounts[t],
			Burn:                   txData.Burns[t],
			Data:                   txData.Data[t],
		}

		if !bytes.Equal(txData.Assets[t], config_coins.NATIVE_ASSET_FULL) {
			transfers[t].FeeRate = txData.Fees[t].Rate
			transfers[t].FeeLeadingZeros = txData.Fees[t].LeadingZeros
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
			if accData := txData.Accs[base64.StdEncoding.EncodeToString(ast)][base64.StdEncoding.EncodeToString(addr.PublicKey)]; len(accData) > 0 {
				if acc, err = account.NewAccount(addr.PublicKey, 0, ast); err != nil {
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
			if regData := txData.Regs[base64.StdEncoding.EncodeToString(addr.PublicKey)]; len(regData) > 0 {
				reg = registration.NewRegistration(addr.PublicKey, 0)
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

		if err = addPoint(senderAddr.EncodeAddr()); err != nil {
			return
		}
		if err = addPoint(txData.Recipients[t]); err != nil {
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

		feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Fees))
		for t := range txData.Fees {
			feesFinal[t] = txData.Fees[t].WizardTransactionFee
		}

		tx, err := wizard.CreateZetherTx(transfers, emap, rings, txData.Height, txData.Hash, publicKeyIndexes, feesFinal, false, ctx, func(status string) {
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
