package wizard

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations/transaction_zether_registration"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"strconv"
)

func GetZetherBalance(publicKey []byte, balanceInit *crypto.ElGamal, asset []byte, hasRollover bool, txs []*transaction.Transaction) (*crypto.ElGamal, error) {
	result, err := GetZetherBalanceMultiple([][]byte{publicKey}, []*crypto.ElGamal{balanceInit}, asset, []bool{hasRollover}, txs)
	if err != nil {
		return nil, err
	}
	return result[0], nil
}

func GetZetherBalanceMultiple(publicKeys [][]byte, balancesInit []*crypto.ElGamal, asset []byte, hasRollovers []bool, txs []*transaction.Transaction) ([]*crypto.ElGamal, error) {

	output := make([]*crypto.ElGamal, len(publicKeys))
	for i, publicKey := range publicKeys {

		changed := false
		var balance *crypto.ElGamal

		for _, tx := range txs {
			if tx.Version == transaction_type.TX_ZETHER {
				base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
				for payloadIndex, payload := range base.Payloads {
					if bytes.Equal(payload.Asset, asset) {
						for j, publicKey2 := range base.Bloom.PublicKeyLists[payloadIndex] {
							if bytes.Equal(publicKey, publicKey2) {

								update := true
								if (j%2 == 1) == payload.Parity && hasRollovers[i] { //receiver
									update = false
								}

								if update {

									if balance == nil {
										if balancesInit[i] != nil {
											balance = balancesInit[i]
										} else {
											var acckey crypto.Point
											if err := acckey.DecodeCompressed(publicKey); err != nil {
												return nil, err
											}
											balance = crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
										}
									}

									echanges := crypto.ConstructElGamal(payload.Statement.C[j], payload.Statement.D)
									balance = balance.Add(echanges) // homomorphic addition of changes
									changed = true
								}

								break
							}
						}
					}
				}
			}
		}

		if changed {
			output[i] = balance
		} else if balancesInit[i] != nil {
			output[i] = balancesInit[i]
		}
	}

	return output, nil
}

func InitializeEmap(assets [][]byte) map[string]map[string][]byte {
	emap := make(map[string]map[string][]byte) //initialize all maps
	for i := range assets {
		if emap[string(assets[i])] == nil {
			emap[string(assets[i])] = make(map[string][]byte)
		}
	}
	return emap
}

func signZetherTx(tx *transaction.Transaction, txBase *transaction_zether.TransactionZether, transfers []*WizardZetherTransfer, emap map[string]map[string][]byte, hasRollovers map[string]bool, ringsSenderMembers, ringsRecipientMembers [][]*bn256.G1, myFees []*WizardTransactionFee, publicKeyIndexes map[string]*WizardZetherPublicKeyIndex, ctx context.Context, statusCallback func(string)) (err error) {

	statusCallback("Transaction Signing...")

	publickeylists := make([][]*bn256.G1, len(transfers))
	parities := make([]bool, len(transfers))

	for t, transfer := range transfers {

		secretPoint := new(crypto.BNRed).SetBytes(transfer.SenderPrivateKey)
		sender := crypto.GPoint.ScalarMult(secretPoint).G1()

		var recipientAddr *addresses.Address
		if recipientAddr, err = addresses.DecodeAddr(transfer.Recipient); err != nil {
			return
		}

		var recipientPoint *crypto.Point
		if recipientPoint, err = recipientAddr.GetPoint(); err != nil {
			return
		}
		recipient := recipientPoint.G1()

		if bytes.Equal(sender.EncodeUncompressed(), recipient.EncodeCompressed()) {
			return errors.New("Sender must NOT be the recipient")
		}
		if bytes.Equal(ringsSenderMembers[t][0].EncodeUncompressed(), sender.EncodeCompressed()) {
			return errors.New("Rings[0] must be the sender")
		}
		if bytes.Equal(ringsRecipientMembers[t][0].EncodeUncompressed(), recipient.EncodeCompressed()) {
			return errors.New("Rings[1] must be the recipient")
		}

		witness_indexes := transfer.WitnessIndexes
		publickeylists[t] = make([]*bn256.G1, 0)

		if len(ringsSenderMembers[t]) != len(ringsRecipientMembers[t]) {
			return fmt.Errorf("Ring Sender %d and Ring Recipient %d should have same length", len(ringsSenderMembers[t]), len(ringsRecipientMembers[t]))
		}
		if len(transfer.WitnessIndexes) != len(ringsSenderMembers[t])+len(ringsRecipientMembers[t]) {
			return fmt.Errorf("Payload %d Witness Indexes length %d is invalid %d", t, len(transfer.WitnessIndexes), len(ringsSenderMembers[t])+len(ringsRecipientMembers[t]))
		}

		parities[t] = witness_indexes[0]%2 == 0

		ringSenderIndex := 1
		ringRecipientIndex := 1

		unique := make(map[string]bool)
		for i := range witness_indexes {

			var publicKey *bn256.G1

			if i == witness_indexes[0] {
				publicKey = sender
			} else if i == witness_indexes[1] {
				publicKey = recipient
			} else if (i%2 == 0) == parities[t] { //sender
				publicKey = ringsSenderMembers[t][ringSenderIndex]
				ringSenderIndex++
			} else { //recipient
				publicKey = ringsRecipientMembers[t][ringRecipientIndex]
				ringRecipientIndex++
			}
			publickeylists[t] = append(publickeylists[t], publicKey)
			unique[publicKey.String()] = true
		}

		if len(unique) != len(witness_indexes) {
			return errors.New("Duplicates detected")
		}

	}
	statusCallback("Transaction public keys were shuffled")

	registrations := make([][]*transaction_zether_registration.TransactionZetherDataRegistration, len(publickeylists))
	registrationsAlready := make(map[string]bool)

	unregisteredAccounts := make([]int, len(transfers))
	unregisteredSpendablePublicKey := make([]int, len(transfers))
	emptyAccounts := make([]int, len(transfers))

	for t, publickeylist := range publickeylists {

		registrations[t] = make([]*transaction_zether_registration.TransactionZetherDataRegistration, len(publickeylist))

		for i, publicKeyPoint := range publickeylist {

			publicKey := publicKeyPoint.EncodeCompressed()

			publicKeyIndex := publicKeyIndexes[string(publicKey)]
			if publicKeyIndex == nil {
				return fmt.Errorf("Public Key Index was not specified for ring member %d", i)
			}

			if !publicKeyIndex.Registered && !registrationsAlready[string(publicKey)] {

				registrationsAlready[string(publicKey)] = true
				if len(publicKeyIndex.RegistrationSignature) != cryptography.SignatureSize {
					return fmt.Errorf("Registration Signature is invalid for ring member %d", i)
				}

				registrations[t][i] = &transaction_zether_registration.TransactionZetherDataRegistration{
					transaction_zether_registration.NOT_REGISTERED,
					publicKeyIndex.RegistrationStaked,
					publicKeyIndex.RegistrationSpendPublicKey,
					publicKeyIndex.RegistrationSignature,
				}

				unregisteredAccounts[t]++

				if len(publicKeyIndex.RegistrationSpendPublicKey) > 0 {
					unregisteredSpendablePublicKey[t]++
				}

			} else if emap[string(transfers[t].Asset)][publicKeyPoint.String()] == nil {
				registrations[t][i] = nil //transaction_zether_registration.REGISTERED_EMPTY_ACCOUNT
				emptyAccounts[t] += 1
			} else {
				registrations[t][i] = nil //transaction_zether_registration.REGISTERED_ACCOUNT
			}

			if emap[string(transfers[t].Asset)][publicKeyPoint.String()] == nil {
				balance := crypto.ConstructElGamal(publicKeyPoint, crypto.ElGamal_BASE_G)
				emap[string(transfers[t].Asset)][publicKeyPoint.String()] = balance.Serialize()
			}

		}
	}

	statusCallback("Transaction registrations created")

	payloads := make([]*transaction_zether_payload.TransactionZetherPayload, len(transfers))
	privateKeysForSign := make([]*addresses.PrivateKey, len(transfers))

	spaceExtra := 0

	for t, transfer := range transfers {

		payloads[t] = &transaction_zether_payload.TransactionZetherPayload{
			Parity: parities[t],
		}

		spaceExtra += unregisteredAccounts[t] * (cryptography.PublicKeySize + 3 + cryptography.SignatureSize) //no of new registrations
		spaceExtra += unregisteredSpendablePublicKey[t] * cryptography.PublicKeySize                          //no of new registrations that will store spend public keys
		spaceExtra += (unregisteredAccounts[t] + emptyAccounts[t]) * (cryptography.PublicKeySize + 1 + 66)    // no of new accounts

		if transfers[t].PayloadExtra == nil {
			payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_TRANSFER
		} else {

			switch payloadExtra := transfers[t].PayloadExtra.(type) {
			case *WizardZetherPayloadExtraStakingReward:
				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_STAKING_REWARD

				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStakingReward{
					Reward:                            payloadExtra.Reward,
					TemporaryAccountRegistrationIndex: uint64(transfer.WitnessIndexes[0]),
				}

				//space extra is 0
			case *WizardZetherPayloadExtraStaking:
				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_STAKING

				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraStaking{}

			case *WizardZetherPayloadExtraAssetCreate:

				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_ASSET_CREATE
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate{nil,
					payloadExtra.Asset,
				}

				spaceExtra += config_coins.ASSET_LENGTH + len(helpers.SerializeToBytes(payloadExtra.Asset))

			case *WizardZetherPayloadExtraSpend:

				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_SPEND

				if privateKeysForSign[t], err = addresses.NewPrivateKey(transfer.SenderSpendPrivateKey); err != nil {
					return
				}
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraSpend{nil,
					privateKeysForSign[t].GeneratePublicKeyPoint(),
					nil,
				}

			case *WizardZetherPayloadExtraAssetSupplyIncrease:
				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE
				if privateKeysForSign[t], err = addresses.NewPrivateKey(payloadExtra.AssetSupplyPrivateKey); err != nil {
					return
				}
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease{nil,
					payloadExtra.AssetId,
					payloadExtra.ReceiverPublicKey,
					payloadExtra.Value,
					privateKeysForSign[t].GeneratePublicKey(),
					helpers.EmptyBytes(cryptography.SignatureSize),
				}

				spaceExtra += 1 + len(payloadExtra.ReceiverPublicKey) + 66

			case *WizardZetherPayloadExtraPlainAccountFund:
				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_PLAIN_ACCOUNT_FUND
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraPlainAccountFund{
					PlainAccountPublicKey: payloadExtra.PlainAccountPublicKey,
				}
			case *WizardZetherPayloadExtraConditionalPayment:
				payloads[t].PayloadScript = transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraConditionalPayment{
					nil,
					payloadExtra.Deadline,
					payloadExtra.DefaultResolution,
					payloadExtra.Threshold,
					payloadExtra.MultisigPublicKeys,
				}
			default:
				return errors.New("Invalid payload")
			}
		}

	}
	tx.SpaceExtra = uint64(spaceExtra)

	var witness_list []crypto.Witness
	sender_secrets := make([]*big.Int, len(transfers))

	otherFee := uint64(0)
	for t, transfer := range transfers {

		select {
		case <-ctx.Done():
			return errors.New("Suspended")
		default:
		}

		var C, CLn, CRn []*bn256.G1
		var D bn256.G1

		publickeylist := publickeylists[t]
		witness_index := transfers[t].WitnessIndexes
		ringSize := len(witness_index)

		var senderKey *addresses.PrivateKey
		if senderKey, err = addresses.NewPrivateKey(transfer.SenderPrivateKey); err != nil {
			return
		}
		secretPoint := new(crypto.BNRed).SetBytes(senderKey.Key)
		sender := crypto.GPoint.ScalarMult(secretPoint).G1()
		sender_secrets[t] = secretPoint.BigInt()

		//  fmt.Printf("len of publickeylist  %d \n", len(publickeylist))

		//  revealing r will disclose the amount and the sender and receiver and separate anonymous ring members
		// calculate r deterministically, so its different every transaction, in emergency it can be given to other, and still will not allows key attacks
		rinputs := append([]byte{}, helpers.CloneBytes(txBase.ChainKernelHash)...)
		for i := range publickeylist {
			rinputs = append(rinputs, publickeylist[i].EncodeCompressed()...)
		}
		rencrypted := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(append([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT), rinputs...))), sender_secrets[t])
		r := crypto.ReducedHash(rencrypted.EncodeCompressed())

		payload := payloads[t]
		payload.Asset = transfers[t].Asset
		payload.BurnValue = transfers[t].Burn

		value := transfers[t].Amount
		burn_value := transfers[t].Burn

		//whisper the value to the sender
		if payload.PayloadScript != transaction_zether_payload_script.SCRIPT_STAKING && payload.PayloadScript != transaction_zether_payload_script.SCRIPT_STAKING_REWARD {
			v2 := crypto.ReducedHash(new(bn256.G1).ScalarMult(publickeylist[witness_index[0]], r).EncodeCompressed())
			v2 = new(big.Int).Add(v2, new(big.Int).SetUint64(value))
			v2Proof := new(big.Int).Mod(v2, bn256.Order)
			payload.WhisperSender = crypto.ConvertBigIntToByte(v2Proof)

			//whisper the value to the recipient
			v1 := crypto.ReducedHash(new(bn256.G1).ScalarMult(publickeylist[witness_index[1]], r).EncodeCompressed())
			v1 = new(big.Int).Add(v1, new(big.Int).SetUint64(value))
			v1proof := new(big.Int).Mod(v1, bn256.Order)
			payload.WhisperRecipient = crypto.ConvertBigIntToByte(v1proof)
		}

		dataFinal := transfer.Data.Data
		payload.DataVersion = transfer.Data.getDataVersion()

		dataLength := len(dataFinal)
		if payload.DataVersion == transaction_data.TX_DATA_NONE {
			dataLength = 0
		} else if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataLength += helpers.BytesLengthSerialized(uint64(len(dataFinal)))
		} else if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
			dataLength = transaction_zether_payload.PAYLOAD_LIMIT
		}

		m := int(math.Log2(float64(ringSize)))

		extraBytes := helpers.BytesLengthSerialized(uint64(payload.PayloadScript)) + helpers.BytesLengthSerialized(payload.BurnValue) //PayloadScript + Burn
		if bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {                                                               //Asset Length
			extraBytes += 1
		} else {
			extraBytes += 1 + len(payload.Asset)
		}
		extraBytes += ringSize * 1                                               //registrations length
		extraBytes += unregisteredAccounts[t] * (1 + cryptography.SignatureSize) //1 byte if it is staked
		extraBytes += 1 + dataLength                                             //dataVersion + data
		extraBytes += ringSize*33*2 + (ringSize-emptyAccounts[t])*33*2 + 33 + 1  //statement
		if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) {
			extraBytes += helpers.BytesLengthSerialized(transfers[t].FeeRate) + 1 //feeRate + FeeLeadingZeros
		}
		extraBytes += 33*(21+m*8) + 32*(10) //proof arrays + proof data
		extraBytes += 2 * m * 32            //proof field array

		if payload.Extra != nil {
			extraBytes += len(transaction_zether_payload_extra.SerializeToBytes(payload.Extra, true))
		}

		extraBytes += len(payload.WhisperRecipient)
		extraBytes += len(payload.WhisperSender)

		fee := setFee(tx, extraBytes, myFees[t].Clone(), t == 0) + otherFee
		otherFee = 0

		statusCallback("Transaction Set fee")

		if !bytes.Equal(transfers[t].Asset, config_coins.NATIVE_ASSET_FULL) {
			payload.FeeRate = transfers[t].FeeRate
			payload.FeeLeadingZeros = transfers[t].FeeLeadingZeros
		}

		if payload.PayloadScript == transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT {
			otherFee = fee
			fee = 0
			payload.FeeRate = 0
			payload.FeeLeadingZeros = 0
		}

		//fake balance
		if payload.PayloadScript == transaction_zether_payload_script.SCRIPT_STAKING_REWARD {

			transfer.SenderDecryptedBalance = value + fee + burn_value

			balance := crypto.ConstructElGamal(sender, crypto.ElGamal_BASE_G)
			balance = balance.Plus(new(big.Int).SetUint64(transfer.SenderDecryptedBalance))

			emap[string(transfer.Asset)][sender.String()] = balance.Serialize()
		}

		// Lots of ToDo for this, enables satisfying lots of  other things
		ebalances_list := make([]*crypto.ElGamal, ringSize)
		for i := range witness_index {
			//in case it is a receiver, it has empty balance
			if (i%2 == 1) == parities[t] || (payload.PayloadScript == transaction_zether_payload_script.SCRIPT_STAKING_REWARD && i != witness_index[0]) {
				ebalances_list[i] = crypto.ConstructElGamal(publickeylist[i], crypto.ElGamal_BASE_G)
			} else {
				var pt *crypto.ElGamal
				if pt, err = new(crypto.ElGamal).Deserialize(emap[string(transfer.Asset)][publickeylist[i].String()]); err != nil {
					return
				}
				ebalances_list[i] = pt
			}
		}

		for i := range publickeylist { // setup commitments
			var x bn256.G1
			switch {
			case i == witness_index[0]:
				x.ScalarMult(crypto.G, new(big.Int).SetInt64(0-int64(value)-int64(fee)-int64(burn_value))) // decrease senders balance
				//fmt.Printf("sender %s \n", x.String())
			case i == witness_index[1]:
				x.ScalarMult(crypto.G, new(big.Int).SetInt64(int64(value))) // increase receiver's balance
				//fmt.Printf("receiver %s \n", x.String())

				// lets encrypt the payment id, it's simple, we XOR the paymentID
				var shared_key []byte
				if shared_key, err = crypto.GenerateSharedSecret(r, publickeylist[i]); err != nil {
					return
				}

				// we must obfuscate it for non-client call
				if len(publickeylist) > config.TRANSACTIONS_ZETHER_RING_MAX {
					return errors.New("currently we do not support ring size > 256")
				}

				if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
					if len(dataFinal) > transaction_zether_payload.PAYLOAD_LIMIT {
						return errors.New("Data final exceeds")
					}
					dataFinal = append(dataFinal, make([]byte, transaction_zether_payload.PAYLOAD_LIMIT-len(dataFinal))...)
					payload.Data = append([]byte{}, dataFinal...)

					// make sure used data encryption is optional, just in case we would like to play together with ring members
					if err = crypto.EncryptDecryptUserData(cryptography.SHA3(append(shared_key[:], publickeylist[i].EncodeCompressed()...)), payload.Data); err != nil {
						return
					}
				} else if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
					if len(dataFinal) > config.TRANSACTIONS_MAX_DATA_LENGTH {
						return errors.New("Data final exceeds")
					}
					payload.Data = dataFinal
				}

			default:
				x.ScalarMult(crypto.G, new(big.Int).SetInt64(0))
			}

			x.Add(new(bn256.G1).Set(&x), new(bn256.G1).ScalarMult(publickeylist[i], r)) // hide all commitments behind r
			C = append(C, &x)
		}
		D.ScalarMult(crypto.G, r)

		//fmt.Printf("t %d publickeylist %d\n", t, len(publickeylist))
		for i := range publickeylist {

			ebalance := ebalances_list[i]

			CLn = append(CLn, new(bn256.G1).Add(ebalance.Left, C[i]))
			CRn = append(CRn, new(bn256.G1).Add(ebalance.Right, &D))
		}

		// decode balance now
		var pt *crypto.ElGamal
		if pt, err = new(crypto.ElGamal).Deserialize(emap[string(transfer.Asset)][sender.String()]); err != nil {
			return
		}

		statusCallback("Homomorphic balance Decrypting...")

		var balance uint64
		if balance, err = senderKey.DecryptBalance(pt, true, transfer.SenderDecryptedBalance, ctx, statusCallback); err != nil {
			return
		}
		transfer.SenderDecryptedBalance = balance //let's update it for the next

		statusCallback("Homomorphic balance Decrypted")

		// time for bullets-sigma
		statement := GenerateStatement(CLn, CRn, publickeylist, C, &D, fee) // generate statement

		statement.RingSize = len(publickeylist)

		witness := GenerateWitness(sender_secrets[t], r, value, balance-value-fee-burn_value, witness_index)

		witness_list = append(witness_list, witness)

		// this goes to proof.u

		//Print(statement, witness)
		payload.Registrations = &transaction_zether_registrations.TransactionZetherDataRegistrations{
			registrations[t],
		}
		payload.Statement = &statement

		// get ready for another round by internal processing of state
		if payload.PayloadScript != transaction_zether_payload_script.SCRIPT_STAKING {
			for i := range publickeylist {

				update := true
				if (i%2 == 0) == payload.Parity { //sender

				} else { //receiver
					if (bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) && hasRollovers[publickeylist[i].String()]) ||
						payload.PayloadScript == transaction_zether_payload_script.SCRIPT_CONDITIONAL_PAYMENT {
						update = false
					}
				}
				if update {
					var encryptedBalance *crypto.ElGamal
					if encryptedBalance, err = new(crypto.ElGamal).Deserialize(emap[string(transfer.Asset)][publickeylist[i].String()]); err != nil {
						return
					}
					echanges := crypto.ConstructElGamal(statement.C[i], statement.D)
					encryptedBalance = encryptedBalance.Add(echanges) // homomorphic addition of changes

					emap[string(transfer.Asset)][publickeylist[i].String()] = encryptedBalance.Serialize() // reserialize and store
				}
			}
		}

	}
	txBase.Payloads = payloads
	statusCallback("Transaction Zether Statements created")

	select {
	case <-ctx.Done():
		return errors.New("Suspended")
	default:
	}

	assetMap := map[string]int{}
	assetIndexes := make([]int, len(transfers))
	for t := range transfers {
		assetIndexes[t] = assetMap[string(txBase.Payloads[t].Asset)]
		assetMap[string(txBase.Payloads[t].Asset)]++
	}

	//make it concurrent
	proofsCn := make([]chan *crypto.Proof, len(transfers))

	statusCallback(fmt.Sprintf("Generating zero knowledge proofs... "))

	hash := tx.GetHashSigningManually()
	for i := range transfers {
		proofsCn[i] = make(chan *crypto.Proof)
		go func(t int) {

			// the u is dependent on roothash,SCID and counter ( counter is dynamic and depends on order of assets)
			uinput := append([]byte(config.PROTOCOL_CRYPTOPGRAPHY_CONSTANT), txBase.ChainKernelHash[:]...)
			uinput = append(uinput, txBase.Payloads[t].Asset[:]...)
			uinput = append(uinput, strconv.Itoa(assetIndexes[t])...)

			u := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(uinput)), sender_secrets[t])

			proof, proofErr := crypto.GenerateProof(txBase.Payloads[t].Asset, assetIndexes[t], txBase.ChainKernelHash, txBase.Payloads[t].Statement, &witness_list[t], u, hash, txBase.Payloads[t].BurnValue)
			if proofErr != nil {
				proofsCn[t] <- nil
				return
			}

			txBase.Payloads[t].Proof = proof

			proofsCn[t] <- proof
		}(i)
	}

	//let's wait for all of them
	for t := range transfers {
		proof := <-proofsCn[t]
		if proof == nil {
			return errors.New("Error generating zk proofs")
		}
	}

	for t := range transfers {
		if privateKeysForSign[t] != nil {

			var signature []byte
			if signature, err = privateKeysForSign[t].Sign(tx.SerializeForSigning()); err != nil {
				return
			}

			switch txBase.Payloads[t].PayloadScript {
			case transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE:
				txBase.Payloads[t].Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease).AssetSignature = signature
			case transaction_zether_payload_script.SCRIPT_SPEND:
				txBase.Payloads[t].Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraSpend).SenderSpendSignature = signature
			}

		}
	}

	statusCallback("Transaction Zether Proofs generated")
	return
}

func CreateZetherTx(transfers []*WizardZetherTransfer, emap map[string]map[string][]byte, hasRollovers map[string]bool, ringsSenderMembers, ringsRecipientMembers [][]*bn256.G1, chainHeight uint64, chainKernelHash []byte, publicKeyIndexes map[string]*WizardZetherPublicKeyIndex, fees []*WizardTransactionFee, ctx context.Context, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	for i, transfer := range transfers {
		if transfer.SenderSpendRequired {
			if len(transfer.SenderSpendPrivateKey) != cryptography.PrivateKeySize {
				return nil, fmt.Errorf("SpendPrivateKey is invalid for payload %d", i)
			}
			if transfer.PayloadExtra != nil {
				return nil, fmt.Errorf("Payload %d requires no payload extra as it will be set automatically to Spend extra", i)
			}
			transfer.PayloadExtra = &WizardZetherPayloadExtraSpend{}
		}
	}

	txBase := &transaction_zether.TransactionZether{
		ChainHeight:     chainHeight,
		ChainKernelHash: chainKernelHash,
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_ZETHER,
		TransactionBaseInterface: txBase,
	}

	if err = signZetherTx(tx, txBase, transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, fees, publicKeyIndexes, ctx, statusCallback); err != nil {
		return
	}
	if err = bloomAllTx(tx, statusCallback); err != nil {
		return
	}

	statusCallback("Transaction Created")
	return tx, nil
}

// generate statement
func GenerateStatement(CLn, CRn, publickeylist, C []*bn256.G1, D *bn256.G1, fee uint64) crypto.Statement {
	return crypto.Statement{CLn: CLn, CRn: CRn, Publickeylist: publickeylist, C: C, D: D, Fee: fee}
}

// generate witness
func GenerateWitness(secretkey, r *big.Int, TransferAmount, Balance uint64, index []int) crypto.Witness {
	return crypto.Witness{SecretKey: secretkey, R: r, TransferAmount: TransferAmount, Balance: Balance, Index: index}
}
