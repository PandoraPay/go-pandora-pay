package builds_data

import (
	"bytes"
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
	"pandora-pay/helpers/advanced_buffers"
	"pandora-pay/txs_builder/txs_builder_zether_helper"
	"pandora-pay/txs_builder/wizard"
)

func PrepareData(data []byte) (txData *TransactionsBuilderCreateZetherTxReq, transfers []*wizard.WizardZetherTransfer, emap map[string]map[string][]byte, hasRollovers map[string]bool, ringsSenderMembers, ringsRecipientMembers [][]*bn256.G1, publicKeyIndexes map[string]*wizard.WizardZetherPublicKeyIndex, feesFinal []*wizard.WizardTransactionFee, err error) {

	txScripts := &struct {
		Payloads []*struct {
			PayloadScript transaction_zether_payload_script.PayloadScriptType `json:"payloadScript"`
		}
	}{}

	if err = json.Unmarshal(data, txScripts); err != nil {
		return
	}

	txData = &TransactionsBuilderCreateZetherTxReq{}
	txData.Payloads = make([]*zetherTxDataPayloadBase, len(txScripts.Payloads))

	for t := range txScripts.Payloads {

		txData.Payloads[t] = &zetherTxDataPayloadBase{}

		switch txScripts.Payloads[t].PayloadScript {
		case transaction_zether_payload_script.SCRIPT_TRANSFER:
			txData.Payloads[t].Extra = nil
		case transaction_zether_payload_script.SCRIPT_ASSET_CREATE:
			txData.Payloads[t].Extra = &wizard.WizardZetherPayloadExtraAssetCreate{}
		case transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE:
			txData.Payloads[t].Extra = &wizard.WizardZetherPayloadExtraAssetSupplyIncrease{}
		case transaction_zether_payload_script.SCRIPT_PLAIN_ACCOUNT_FUND:
			txData.Payloads[t].Extra = &wizard.WizardZetherPayloadExtraPlainAccountFund{}
		case transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT:
			txData.Payloads[t].Extra = &wizard.WizardZetherPayloadExtraConditionalPayment{}
		default:
			err = errors.New("Invalid PayloadScriptType")
			return
		}

	}

	if err = json.Unmarshal(data, txData); err != nil {
		return
	}

	transfers = make([]*wizard.WizardZetherTransfer, len(txData.Payloads))
	ringsSenderMembers = make([][]*bn256.G1, len(txData.Payloads))
	ringsRecipientMembers = make([][]*bn256.G1, len(txData.Payloads))

	publicKeyIndexes = make(map[string]*wizard.WizardZetherPublicKeyIndex)
	hasRollovers = make(map[string]bool)

	sendAssets := make([][]byte, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		sendAssets[t] = payload.Asset
	}
	emap = wizard.InitializeEmap(sendAssets)

	senderRingMembers := make([][]string, len(txData.Payloads))
	recipientRingMembers := make([][]string, len(txData.Payloads))

	txDataPayloads := &txs_builder_zether_helper.TxsBuilderZetherTxDataBase{
		Payloads: make([]*txs_builder_zether_helper.TxsBuilderZetherTxPayloadBase, len(txData.Payloads)),
	}
	for i := range txDataPayloads.Payloads {
		txDataPayloads.Payloads[i] = &txData.Payloads[i].TxsBuilderZetherTxPayloadBase
	}

	for t, payload := range txData.Payloads {

		transfers[t] = &wizard.WizardZetherTransfer{
			Asset:                  payload.Asset,
			SenderPrivateKey:       payload.SenderData.PrivateKey,
			SenderDecryptedBalance: payload.SenderData.DecryptedBalance,
			Recipient:              payload.Recipient,
			Amount:                 payload.Amount,
			Burn:                   payload.Burn,
			Data:                   payload.Data,
			PayloadExtra:           payload.Extra,
		}

		if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
			transfers[t].FeeRate = payload.Fees.Rate
			transfers[t].FeeLeadingZeros = payload.Fees.LeadingZeros
		}

		uniqueMap := make(map[string]bool)
		ringSender := make([]*bn256.G1, 0)
		ringRecipient := make([]*bn256.G1, 0)

		addPoint := func(address string, sender, isSender bool) (err error) {

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

			var reg *registration.Registration
			if regData := txData.Regs[base64.StdEncoding.EncodeToString(addr.PublicKey)]; len(regData) > 0 {
				reg = registration.NewRegistration(addr.PublicKey, 0)
				if err = reg.Deserialize(advanced_buffers.NewBufferReader(regData)); err != nil {
					return
				}
			}

			if emap[string(payload.Asset)][p.G1().String()] == nil {

				var acc *account.Account
				if accData := txData.Accs[base64.StdEncoding.EncodeToString(payload.Asset)][base64.StdEncoding.EncodeToString(addr.PublicKey)]; len(accData) > 0 {
					if acc, err = account.NewAccount(addr.PublicKey, 0, payload.Asset); err != nil {
						return
					}
					if err = acc.Deserialize(advanced_buffers.NewBufferReader(accData)); err != nil {
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

				hasRollovers[p.G1().String()] = reg != nil && reg.Staked

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
						publicKeyIndex.RegistrationStaked = addr.Staked
						publicKeyIndex.RegistrationSpendPublicKey = addr.SpendPublicKey
						publicKeyIndex.RegistrationSignature = addr.Registration
					}
				}
			}

			if isSender { //sender
				if reg != nil && len(reg.SpendPublicKey) > 0 && payload.Extra == nil {
					transfers[t].SenderSpendRequired = true
					if payload.SenderData.SpendPrivateKey == nil {
						return errors.New("Spend Private Key is missing")
					}

					var spendPrivateKey *addresses.PrivateKey
					if spendPrivateKey, err = addresses.NewPrivateKey(payload.SenderData.SpendPrivateKey); err != nil {
						return
					}

					spendPublicKey := spendPrivateKey.GeneratePublicKey()
					if !bytes.Equal(spendPublicKey, reg.SpendPublicKey) {
						return errors.New("Wallet Spend Public Key is not matching")
					}
					transfers[t].SenderSpendPrivateKey = payload.SenderData.SpendPrivateKey
				}
			}

			if sender {
				ringSender = append(ringSender, p.G1())
			} else {
				ringRecipient = append(ringRecipient, p.G1())
			}

			return
		}

		for i, ringMember := range payload.SenderRingMembers {
			if err = addPoint(ringMember, true, i == 0); err != nil {
				return
			}
		}
		for _, ringMember := range payload.RecipientRingMembers {
			if err = addPoint(ringMember, false, false); err != nil {
				return
			}
		}

		ringsSenderMembers[t] = ringSender
		ringsRecipientMembers[t] = ringRecipient

		txs_builder_zether_helper.InitRing(t, senderRingMembers, recipientRingMembers, &payload.TxsBuilderZetherTxPayloadBase)

		senderRingMembers[t] = payload.SenderRingMembers
		recipientRingMembers[t] = payload.RecipientRingMembers

		if err = txs_builder_zether_helper.ProcessRing(t, senderRingMembers, recipientRingMembers, txDataPayloads); err != nil {
			return
		}

		transfers[t].WitnessIndexes = payload.WitnessIndexes
	}

	feesFinal = make([]*wizard.WizardTransactionFee, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		feesFinal[t] = payload.Fees.WizardTransactionFee
	}

	return
}
