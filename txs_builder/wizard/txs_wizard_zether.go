package wizard

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/dpos"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
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

func InitializeEmap(assets [][]byte) map[string]map[string][]byte {
	emap := make(map[string]map[string][]byte) //initialize all maps
	for i := range assets {
		if emap[string(assets[i])] == nil {
			emap[string(assets[i])] = make(map[string][]byte)
		}
	}
	return emap
}

func signZetherTx(tx *transaction.Transaction, txBase *transaction_zether.TransactionZether, transfers []*WizardZetherTransfer, emap map[string]map[string][]byte, rings [][]*bn256.G1, myFees []*WizardTransactionFee, chainHeight uint64, chainHash []byte, publicKeyIndexes map[string]*WizardZetherPublicKeyIndex, ctx context.Context, statusCallback func(string)) (err error) {

	statusCallback("Transaction Signing...")

	publickeylists := make([][]*bn256.G1, len(transfers))
	witness_indexes := make([][]int, len(transfers))

	for t, transfer := range transfers {

		senderKey := &addresses.PrivateKey{Key: transfer.From}
		secretPoint := new(crypto.BNRed).SetBytes(senderKey.Key)
		sender := crypto.GPoint.ScalarMult(secretPoint).G1()

		var receiver_addr *addresses.Address
		if receiver_addr, err = addresses.DecodeAddr(transfer.Destination); err != nil {
			return
		}

		var receiverPoint *crypto.Point
		if receiverPoint, err = receiver_addr.GetPoint(); err != nil {
			return
		}
		receiver := receiverPoint.G1()

		if bytes.Equal(sender.EncodeUncompressed(), receiver.EncodeCompressed()) {
			return errors.New("Sender must be the receiver")
		}
		if bytes.Equal(rings[t][0].EncodeUncompressed(), sender.EncodeCompressed()) {
			return errors.New("Rings[0] must be the sender")
		}
		if bytes.Equal(rings[t][1].EncodeUncompressed(), receiver.EncodeCompressed()) {
			return errors.New("Rings[1] must be the receiver")
		}

		witness_indexes[t] = helpers.ShuffleArray_for_Zether(len(rings[t]))
		anonset_publickeys := rings[t][2:]
		publickeylists[t] = make([]*bn256.G1, 0)

		unique := make(map[string]bool)
		for i := range witness_indexes[t] {

			var publicKey *bn256.G1
			switch i {
			case witness_indexes[t][0]:
				publicKey = sender
			case witness_indexes[t][1]:
				publicKey = receiver
			default:
				publicKey = anonset_publickeys[0]
				anonset_publickeys = anonset_publickeys[1:]
			}
			publickeylists[t] = append(publickeylists[t], publicKey)
			unique[string(publicKey.EncodeCompressed())] = true
		}
		if len(unique) != len(rings[t]) {
			return errors.New("Duplicates detected")
		}

	}
	statusCallback("Transaction public keys were shuffled")

	registrations := make([][]*transaction_zether_registration.TransactionZetherDataRegistration, len(publickeylists))
	registrationsAlready := make(map[string]bool)

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
					publicKeyIndex.RegistrationSignature,
				}

			} else if emap[string(transfers[t].Asset)][publicKeyPoint.String()] == nil {

				registrations[t][i] = &transaction_zether_registration.TransactionZetherDataRegistration{
					transaction_zether_registration.REGISTERED_EMPTY_ACCOUNT,
					nil,
				}

			} else {
				registrations[t][i] = &transaction_zether_registration.TransactionZetherDataRegistration{
					transaction_zether_registration.REGISTERED_ACCOUNT,
					nil,
				}
			}

			if emap[string(transfers[t].Asset)][publicKeyPoint.String()] == nil {
				var acckey crypto.Point
				if err = acckey.DecodeCompressed(publicKeyPoint.EncodeCompressed()); err != nil {
					return
				}
				balance := crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
				emap[string(transfers[t].Asset)][publicKeyPoint.String()] = balance.Serialize()
			}

		}
	}
	statusCallback("Transaction registrations created")

	payloads := make([]*transaction_zether_payload.TransactionZetherPayload, len(transfers))
	privateKeysForSign := make([]*addresses.PrivateKey, len(transfers))

	spaceExtra := 0

	for t, transfer := range transfers {

		publickeylist := publickeylists[t]
		senderKey := &addresses.PrivateKey{Key: transfer.From}

		payloads[t] = &transaction_zether_payload.TransactionZetherPayload{}

		var unregisteredAccounts, emptyAccounts int
		for _, reg := range registrations[t] {
			if reg.RegistrationType == transaction_zether_registration.NOT_REGISTERED {
				unregisteredAccounts += 1
			}
			if reg.RegistrationType == transaction_zether_registration.REGISTERED_EMPTY_ACCOUNT {
				emptyAccounts += 1
			}
		}

		spaceExtra += unregisteredAccounts * (cryptography.PublicKeySize + 1 + cryptography.SignatureSize)
		spaceExtra += (unregisteredAccounts + emptyAccounts) * (cryptography.PublicKeySize + 1 + 66)

		if transfers[t].PayloadExtra == nil {
			payloads[t].PayloadScript = transaction_zether_payload.SCRIPT_TRANSFER
		} else {

			switch payloadExtra := transfers[t].PayloadExtra.(type) {
			case *WizardZetherPayloadExtraClaim:
				payloads[t].PayloadScript = transaction_zether_payload.SCRIPT_CLAIM

				var registrationIndex byte

				senderPublicKey := senderKey.GeneratePublicKey()
				for i := range registrations[t] {
					if bytes.Equal(publickeylist[i].EncodeCompressed(), senderPublicKey) {
						registrationIndex = byte(i)
						break
					}
				}

				key := &addresses.PrivateKey{Key: payloadExtra.DelegatePrivateKey}
				delegatePublicKey := key.GeneratePublicKey()
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraClaim{
					DelegatePublicKey:           delegatePublicKey,
					DelegatedStakingClaimAmount: transfers[t].Amount,
					RegistrationIndex:           registrationIndex,
					DelegateSignature:           helpers.EmptyBytes(cryptography.SignatureSize),
				}

				privateKeysForSign[t] = key

				//space extra is 0

			case *WizardZetherPayloadExtraDelegateStake:
				payloads[t].PayloadScript = transaction_zether_payload.SCRIPT_DELEGATE_STAKE

				blankSignature := []byte{}
				if payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
					key := &addresses.PrivateKey{Key: payloadExtra.DelegatePrivateKey}
					if bytes.Equal(key.GeneratePublicKey(), payloadExtra.DelegatePublicKey) == false {
						return errors.New("delegatePrivateKey is not matching delegatePublicKey")
					}

					privateKeysForSign[t] = key
					blankSignature = helpers.EmptyBytes(cryptography.SignatureSize)
				}

				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake{
					DelegatePublicKey:      payloadExtra.DelegatePublicKey,
					ConvertToUnclaimed:     payloadExtra.ConvertToUnclaimed,
					DelegatedStakingUpdate: payloadExtra.DelegatedStakingUpdate,
					DelegateSignature:      blankSignature,
				}

				if payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
					spaceExtra += len(payloadExtra.DelegatePublicKey) //not required when account exists
					spaceExtra += len(payloadExtra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey)
					spaceExtra += helpers.BytesLengthSerialized(payloadExtra.DelegatedStakingUpdate.DelegatedStakingNewFee)
				}

				if payloadExtra.ConvertToUnclaimed {
					spaceExtra += binary.MaxVarintLen64
				} else {
					spaceExtra += len(helpers.SerializeToBytes(&dpos.DelegatedStakePending{nil, transfers[t].Burn, chainHeight + 100, dpos.DelegatedStakePendingStake}))
					spaceExtra += cryptography.PublicKeySize //key
				}

			case *WizardZetherPayloadExtraAssetCreate:
				payloads[t].PayloadScript = transaction_zether_payload.SCRIPT_ASSET_CREATE
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate{
					Asset: payloadExtra.Asset,
				}

				spaceExtra += config_coins.ASSET_LENGTH + len(helpers.SerializeToBytes(payloadExtra.Asset))

			case *WizardZetherPayloadExtraAssetSupplyIncrease:
				payloads[t].PayloadScript = transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE
				privateKeysForSign[t] = &addresses.PrivateKey{Key: payloadExtra.AssetSupplyPrivateKey}
				payloads[t].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease{
					AssetId:              payloadExtra.AssetId,
					ReceiverPublicKey:    payloadExtra.ReceiverPublicKey,
					Value:                payloadExtra.Value,
					AssetSupplyPublicKey: privateKeysForSign[t].GeneratePublicKey(),
					AssetSignature:       helpers.EmptyBytes(cryptography.SignatureSize),
				}

				spaceExtra += binary.MaxVarintLen64

			default:
				return errors.New("Invalid payload")
			}
		}

	}
	tx.SpaceExtra = uint64(spaceExtra)

	var witness_list []crypto.Witness
	for t, transfer := range transfers {

		select {
		case <-ctx.Done():
			return errors.New("Suspended")
		default:
		}

		var C, CLn, CRn []*bn256.G1
		var D bn256.G1

		publickeylist := publickeylists[t]
		witness_index := witness_indexes[t]

		senderKey := &addresses.PrivateKey{Key: transfer.From}
		secretPoint := new(crypto.BNRed).SetBytes(senderKey.Key)
		sender := crypto.GPoint.ScalarMult(secretPoint).G1()
		sender_secret := secretPoint.BigInt()

		//  fmt.Printf("len of publickeylist  %d \n", len(publickeylist))

		//  revealing r will disclose the amount and the sender and receiver and separate anonymous ring members
		// calculate r deterministically, so its different every transaction, in emergency it can be given to other, and still will not allows key attacks
		rinputs := append([]byte{}, helpers.CloneBytes(chainHash)...)
		for i := range publickeylist {
			rinputs = append(rinputs, publickeylist[i].EncodeCompressed()...)
		}
		rencrypted := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(append([]byte(crypto.PROTOCOL_CONSTANT), rinputs...))), sender_secret)
		r := crypto.ReducedHash(rencrypted.EncodeCompressed())

		payload := payloads[t]
		payload.Asset = transfers[t].Asset
		payload.BurnValue = transfers[t].Burn

		value := transfers[t].Amount
		burn_value := transfers[t].Burn

		dataFinal := transfer.Data.Data
		payload.DataVersion = transfer.Data.getDataVersion()

		dataLength := len(dataFinal)
		if payload.DataVersion == transaction_data.TX_DATA_NONE {
			dataLength = 0
		} else if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataLength += helpers.BytesLengthSerialized(uint64(len(dataFinal)))
		} else if payload.DataVersion == transaction_data.TX_DATA_ENCRYPTED {
			dataLength = transaction_zether_payload.PAYLOAD0_LIMIT
		}

		m := int(math.Log2(float64(len(rings[t]))))

		extraBytes := 1 + len(payload.Asset) + helpers.BytesLengthSerialized(payload.BurnValue)
		extraBytes += 1 + cryptography.SignatureSize + 1 //registrations length
		extraBytes += 1 + dataLength                     //dataVersion + data
		extraBytes += len(rings[t])*33*4 + 33 + 1        // statement
		extraBytes += 33*(21+m*8) + 32*(10)              //proof arrays + proof data
		extraBytes += 2 * m * 32                         //proof field array

		if payload.Extra != nil {
			extraBytes += len(transaction_zether_payload_extra.SerializeToBytes(payload.Extra, true))
		}

		fee := setFee(tx, extraBytes, myFees[t].Clone(), t == 0)

		statusCallback("Transaction Set fee")

		if !bytes.Equal(transfers[t].Asset, config_coins.NATIVE_ASSET_FULL) {
			payload.FeeRate = transfers[t].FeeRate
			payload.FeeLeadingZeros = transfers[t].FeeLeadingZeros

			fee = (fee * helpers.Pow10(transfers[t].FeeLeadingZeros)) / transfers[t].FeeRate
			statusCallback("Transaction Set conversion fee")
		}

		//fake balance
		if payload.PayloadScript == transaction_zether_payload.SCRIPT_CLAIM {

			transfer.FromBalanceDecoded = value + fee + burn_value

			var acckey crypto.Point
			if err = acckey.DecodeCompressed(senderKey.GeneratePublicKey()); err != nil {
				return
			}
			balance := crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)
			balance = balance.Plus(new(big.Int).SetUint64(transfer.FromBalanceDecoded))

			emap[string(transfer.Asset)][sender.String()] = balance.Serialize()
		}

		// Lots of ToDo for this, enables satisfying lots of  other things
		ebalances_list := make([]*crypto.ElGamal, len(rings[t]))
		for i := range witness_index {
			var pt *crypto.ElGamal
			if pt, err = new(crypto.ElGamal).Deserialize(emap[string(transfer.Asset)][publickeylist[i].String()]); err != nil {
				return
			}
			ebalances_list[i] = pt
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
					if len(dataFinal) > transaction_zether_payload.PAYLOAD0_LIMIT {
						return errors.New("Data final exceeds")
					}
					dataFinal = append(dataFinal, make([]byte, transaction_zether_payload.PAYLOAD0_LIMIT-len(dataFinal))...)
					payload.Data = append([]byte{byte(uint(witness_index[0]))}, dataFinal...)

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

			var ll, rr bn256.G1

			ll.Add(ebalance.Left, C[i])
			CLn = append(CLn, &ll)

			rr.Add(ebalance.Right, &D)
			CRn = append(CRn, &rr)
		}

		// decode balance now
		var pt *crypto.ElGamal
		if pt, err = new(crypto.ElGamal).Deserialize(emap[string(transfer.Asset)][sender.String()]); err != nil {
			return
		}

		statusCallback("Homomorphic balance Decoding...")

		var balance uint64
		if balance, err = senderKey.DecodeBalance(pt, transfer.FromBalanceDecoded, ctx, statusCallback); err != nil {
			return
		}
		transfer.FromBalanceDecoded = balance //let's update it for the next

		statusCallback("Homomorphic balance Decoded")

		// time for bullets-sigma
		statement := GenerateStatement(CLn, CRn, publickeylist, C, &D, fee) // generate statement

		statement.RingSize = len(publickeylist)

		witness := GenerateWitness(sender_secret, r, value, balance-value-fee-burn_value, witness_index)

		witness_list = append(witness_list, witness)

		// this goes to proof.u

		//Print(statement, witness)
		payload.Registrations = &transaction_zether_registrations.TransactionZetherDataRegistrations{
			registrations[t],
		}
		payload.Statement = &statement

		// get ready for another round by internal processing of state
		for i := range publickeylist {

			var balance *crypto.ElGamal
			if balance, err = new(crypto.ElGamal).Deserialize(emap[string(transfer.Asset)][publickeylist[i].String()]); err != nil {
				return
			}
			echanges := crypto.ConstructElGamal(statement.C[i], statement.D)

			balance = balance.Add(echanges)                                               // homomorphic addition of changes
			emap[string(transfer.Asset)][publickeylist[i].String()] = balance.Serialize() // reserialize and store
		}

	}
	txBase.Payloads = payloads
	statusCallback("Transaction Zether Statements created")

	senderKey := &addresses.PrivateKey{Key: transfers[0].From}
	sender_secret := new(crypto.BNRed).SetBytes(senderKey.Key).BigInt()

	assetMap := map[string]int{}
	for t := range transfers {

		select {
		case <-ctx.Done():
			return errors.New("Suspended")
		default:
		}

		// the u is dependent on roothash,SCID and counter ( counter is dynamic and depends on order of assets)
		uinput := append([]byte(crypto.PROTOCOL_CONSTANT), chainHash[:]...)
		assetIndex := assetMap[string(txBase.Payloads[t].Asset)]
		uinput = append(uinput, txBase.Payloads[t].Asset[:]...)
		uinput = append(uinput, strconv.Itoa(assetIndex)...)

		u := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(uinput)), sender_secret) // this should be moved to generate proof

		assetMap[string(txBase.Payloads[t].Asset)] = assetMap[string(txBase.Payloads[t].Asset)] + 1

		statusCallback(fmt.Sprintf("Payload %d generating zero knowledge proofs... ", t+1))
		if txBase.Payloads[t].Proof, err = crypto.GenerateProof(txBase.Payloads[t].Asset, assetIndex, txBase.ChainHash, txBase.Payloads[t].Statement, &witness_list[t], u, tx.GetHashSigningManually(), txBase.Payloads[t].BurnValue); err != nil {
			return
		}
	}

	for t := range transfers {
		if privateKeysForSign[t] != nil {

			var signature []byte
			if signature, err = privateKeysForSign[t].Sign(tx.SerializeForSigning()); err != nil {
				return
			}

			switch txBase.Payloads[t].PayloadScript {
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				txBase.Payloads[t].Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake).DelegateSignature = signature
			case transaction_zether_payload.SCRIPT_CLAIM:
				txBase.Payloads[t].Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraClaim).DelegateSignature = signature
			case transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
				txBase.Payloads[t].Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease).AssetSignature = signature
			}

		}
	}

	statusCallback("Transaction Zether Proofs generated")
	return
}

func CreateZetherTx(transfers []*WizardZetherTransfer, emap map[string]map[string][]byte, rings [][]*bn256.G1, chainHeight uint64, chainHash []byte, publicKeyIndexes map[string]*WizardZetherPublicKeyIndex, fees []*WizardTransactionFee, validateTx bool, ctx context.Context, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	txBase := &transaction_zether.TransactionZether{
		ChainHeight: chainHeight,
		ChainHash:   chainHash,
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_ZETHER,
		TransactionBaseInterface: txBase,
	}

	if err = signZetherTx(tx, txBase, transfers, emap, rings, fees, chainHeight, chainHash, publicKeyIndexes, ctx, statusCallback); err != nil {
		return
	}
	if err = bloomAllTx(tx, validateTx, statusCallback); err != nil {
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
