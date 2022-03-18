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
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
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
	SpendPrivateKey  []byte `json:"spendPrivateKey"`
	DecryptedBalance uint64 `json:"decryptedBalance"`
}

type zetherTxDataPayloadBase struct {
	Sender      *zetherTxDataSender                                 `json:"sender"`
	Asset       []byte                                              `json:"asset"`
	Amount      uint64                                              `json:"amount"`
	Recipient   string                                              `json:"recipient"`
	Burn        uint64                                              `json:"burn"`
	RingMembers []string                                            `json:"ringMembers"`
	Data        *wizard.WizardTransactionData                       `json:"data"`
	Fees        *wizard.WizardZetherTransactionFee                  `json:"fees"`
	ScriptType  transaction_zether_payload_script.PayloadScriptType `json:"scriptType"`
	Extra       wizard.WizardZetherPayloadExtra                     `json:"extra"`
}

type zetherTxDataBase struct {
	Payloads          []*zetherTxDataPayloadBase   `json:"payloads"`
	Accs              map[string]map[string][]byte `json:"accs"`
	Regs              map[string][]byte            `json:"regs"`
	ChainKernelHeight uint64                       `json:"chainKernelHeight"`
	ChainKernelHash   []byte                       `json:"chainKernelHash"`
}

func prepareData(txData *zetherTxDataBase) (transfers []*wizard.WizardZetherTransfer, emap map[string]map[string][]byte, hasRollovers map[string]bool, rings [][]*bn256.G1, publicKeyIndexes map[string]*wizard.WizardZetherPublicKeyIndex, err error) {

	transfers = make([]*wizard.WizardZetherTransfer, len(txData.Payloads))
	rings = make([][]*bn256.G1, len(txData.Payloads))
	publicKeyIndexes = make(map[string]*wizard.WizardZetherPublicKeyIndex)
	hasRollovers = make(map[string]bool)

	sendAssets := make([][]byte, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		sendAssets[t] = payload.Asset
	}
	emap = wizard.InitializeEmap(sendAssets)

	for t, payload := range txData.Payloads {

		key := addresses.PrivateKey{Key: payload.Sender.PrivateKey}

		var senderAddr *addresses.Address
		senderAddr, err = key.GenerateAddress(false, nil, txData.Regs[string(key.GeneratePublicKey())] == nil, nil, 0, nil)
		if err != nil {
			return
		}

		transfers[t] = &wizard.WizardZetherTransfer{
			Asset:                  payload.Asset,
			SenderPrivateKey:       payload.Sender.PrivateKey,
			SenderDecryptedBalance: payload.Sender.DecryptedBalance,
			Recipient:              payload.Recipient,
			Amount:                 payload.Amount,
			Burn:                   payload.Burn,
			Data:                   payload.Data,
		}

		if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
			transfers[t].FeeRate = payload.Fees.Rate
			transfers[t].FeeLeadingZeros = payload.Fees.LeadingZeros
		}

		var payloadExtra wizard.WizardZetherPayloadExtra
		switch payload.ScriptType {
		case transaction_zether_payload_script.SCRIPT_TRANSFER:
			payloadExtra = nil
		case transaction_zether_payload_script.SCRIPT_ASSET_CREATE:
			payloadExtra = &wizard.WizardZetherPayloadExtraAssetCreate{}
		case transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE:
			payloadExtra = &wizard.WizardZetherPayloadExtraAssetSupplyIncrease{}
		default:
			err = errors.New("Invalid PayloadScriptType")
			return
		}

		if payloadExtra != nil {
			var data []byte
			if data, err = json.Marshal(payload.Extra); err != nil {
				return
			}
			if err = json.Unmarshal(data, payloadExtra); err != nil {
				return
			}
		}

		transfers[t].PayloadExtra = payloadExtra

		uniqueMap := make(map[string]bool)
		var ring []*bn256.G1

		addPoint := func(address string, isSender bool) (err error) {

			var addr *addresses.Address
			var p *crypto.Point

			if addr, err = addresses.DecodeAddr(address); err != nil {
				return
			}
			if uniqueMap[string(addr.PublicKey)] {
				return errors.New("Ring Member detected twice")
			}
			uniqueMap[string(addr.PublicKey)] = true

			if p, err = addr.GetPoint(); err != nil {
				return
			}

			if emap[string(payload.Asset)][p.G1().String()] == nil {
				var reg *registration.Registration
				if regData := txData.Regs[base64.StdEncoding.EncodeToString(addr.PublicKey)]; len(regData) > 0 {
					reg = registration.NewRegistration(addr.PublicKey, 0)
					if err = reg.Deserialize(helpers.NewBufferReader(regData)); err != nil {
						return
					}
				}

				var acc *account.Account
				if accData := txData.Accs[base64.StdEncoding.EncodeToString(payload.Asset)][base64.StdEncoding.EncodeToString(addr.PublicKey)]; len(accData) > 0 {
					if acc, err = account.NewAccount(addr.PublicKey, 0, payload.Asset); err != nil {
						return
					}
					if err = acc.Deserialize(helpers.NewBufferReader(accData)); err != nil {
						return
					}
					emap[string(payload.Asset)][p.G1().String()] = acc.Balance.Amount.Serialize()
				} else {
					var acckey crypto.Point
					if err = acckey.DecodeCompressed(addr.PublicKey); err != nil {
						return
					}
					emap[string(payload.Asset)][p.G1().String()] = crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G).Serialize()
				}

				hasRollovers[p.G1().String()] = reg != nil && reg.Stakable

				if isSender { //sender
					if reg != nil && len(reg.SpendPublicKey) > 0 && payload.Extra == nil {
						transfers[t].SenderUnstakeRequired = true
						if payload.Sender.SpendPrivateKey == nil {
							return errors.New("Spend Private Key is missing")
						}
						spendPublicKey := (&addresses.PrivateKey{Key: payload.Sender.SpendPrivateKey}).GeneratePublicKey()
						if !bytes.Equal(spendPublicKey, reg.SpendPublicKey) {
							return errors.New("Wallet Spend Public Key is not matching")
						}
						transfers[t].SenderSpendPrivateKey = payload.Sender.SpendPrivateKey
					}
				}

				if publicKeyIndexes[string(addr.PublicKey)] == nil {
					publicKeyIndex := &wizard.WizardZetherPublicKeyIndex{}
					publicKeyIndexes[string(addr.PublicKey)] = publicKeyIndex

					if reg != nil {
						publicKeyIndex.Registered = true
						publicKeyIndex.RegisteredIndex = reg.Index
					} else {
						if len(addr.Registration) == 0 {
							return fmt.Errorf("Signature is missing for %s", addr.EncodeAddr())
						}
						publicKeyIndex.RegistrationStakable = addr.Stakable
						publicKeyIndex.RegistrationSpendPublicKey = addr.SpendPublicKey
						publicKeyIndex.RegistrationSignature = addr.Registration
					}
				}
			}

			ring = append(ring, p.G1())

			return
		}

		if err = addPoint(senderAddr.EncodeAddr(), true); err != nil {
			return
		}
		if err = addPoint(payload.Recipient, false); err != nil {
			return
		}
		for _, ringMember := range payload.RingMembers {
			if err = addPoint(ringMember, false); err != nil {
				return
			}
		}

		rings[t] = ring
		transfers[t].WitnessIndexes = helpers.ShuffleArray_for_Zether(len(ring))

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

		transfers, emap, hasRollovers, rings, publicKeyIndexes, err := prepareData(txData)
		if err != nil {
			return nil, err
		}

		feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Payloads))
		for t, payload := range txData.Payloads {
			feesFinal[t] = payload.Fees.WizardTransactionFee
		}

		tx, err := wizard.CreateZetherTx(transfers, emap, hasRollovers, rings, txData.ChainKernelHeight, txData.ChainKernelHash, publicKeyIndexes, feesFinal, ctx, func(status string) {
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
