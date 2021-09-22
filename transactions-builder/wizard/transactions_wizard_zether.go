package wizard

import (
	"errors"
	"fmt"
	"math/big"
	mathrand "math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type ZetherTransfer struct {
	Token              helpers.HexBytes
	From               []byte //private key
	FromBalanceDecoded uint64
	Destination        string
	Amount             uint64
	Burn               uint64
	Data               *TransactionsWizardData
}

type ZetherPublicKeyIndex struct {
	Registered            bool
	RegisteredIndex       uint64
	RegistrationSignature []byte
}

func CreateZetherTx(transfers []*ZetherTransfer, emap map[string]map[string][]byte, rings [][]*bn256.G1, height uint64, hash []byte, publicKeyIndexes map[string]*ZetherPublicKeyIndex, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	txBase := &transaction_zether.TransactionZether{
		TxScript: transaction_zether.SCRIPT_TRANSFER,
		Height:   height,
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_ZETHER,
		TransactionBaseInterface: txBase,
	}

	registrations := make([]*transaction_zether.TransactionZetherRegistration, 0)
	registrationsAlready := make(map[string]bool)

	index := uint64(0)
	for _, ring := range rings {
		for ringIndex, publicKeyPoint := range ring {

			publicKey := publicKeyPoint.EncodeCompressed()
			index += 1

			if publicKeyIndex := publicKeyIndexes[string(publicKey)]; publicKeyIndex != nil {

				if !publicKeyIndex.Registered && !registrationsAlready[string(publicKey)] {
					registrationsAlready[string(publicKey)] = true
					registrations = append(registrations, &transaction_zether.TransactionZetherRegistration{
						index,
						publicKeyIndex.RegistrationSignature,
					})
				}

			} else {
				return nil, fmt.Errorf("Public Key Index was not specified for ring member %d", ringIndex)
			}

		}
	}
	txBase.Registrations = registrations

	statusCallback("Transaction created")
	var witness_list []crypto.Witness

	for t, transfer := range transfers {

		senderKey := &addresses.PrivateKey{Key: transfer.From}
		secretPoint := new(crypto.BNRed).SetBytes(senderKey.Key)
		sender := crypto.GPoint.ScalarMult(secretPoint).G1()
		sender_secret := secretPoint.BigInt()

		crand := mathrand.New(helpers.NewCryptoRandSource())

		var publickeylist, C, CLn, CRn []*bn256.G1
		var D bn256.G1

		var receiver_addr *addresses.Address
		if receiver_addr, err = addresses.DecodeAddr(transfer.Destination); err != nil {
			return
		}

		var receiverPoint *crypto.Point
		if receiverPoint, err = receiver_addr.GetPoint(); err != nil {
			return
		}
		receiver := receiverPoint.G1()

		var witness_index []int
		for i := 0; i < len(rings[t]); i++ { // todocheck whether this is power of 2 or not
			witness_index = append(witness_index, i)
		}

		//witness_index[3], witness_index[1] = witness_index[1], witness_index[3]
		for {
			crand.Shuffle(len(witness_index), func(i, j int) {
				witness_index[i], witness_index[j] = witness_index[j], witness_index[i]
			})

			// make sure sender and receiver are not both odd or both even
			// sender will always be at  witness_index[0] and receiver will always be at witness_index[1]
			if witness_index[0]%2 != witness_index[1]%2 {
				break
			}
		}

		// Lots of ToDo for this, enables satisfying lots of  other things
		anonset_publickeys := rings[t][2:]
		ebalances_list := make([]*crypto.ElGamal, 0, len(rings[t]))
		for i := range witness_index {

			var data string
			switch i {
			case witness_index[0]:
				publickeylist = append(publickeylist, sender)
				data = sender.String()
			case witness_index[1]:
				publickeylist = append(publickeylist, receiver)
				data = receiver.String()
			default:
				publickeylist = append(publickeylist, anonset_publickeys[0])
				data = anonset_publickeys[0].String()
				anonset_publickeys = anonset_publickeys[1:]
			}

			var pt *crypto.ElGamal
			if pt, err = new(crypto.ElGamal).Deserialize(emap[string(transfers[t].Token)][data]); err != nil {
				return
			}
			ebalances_list = append(ebalances_list, pt)

			// fmt.Printf("adding %d %s  (ring count %d) \n", i,publickeylist[i].String(), len(anonset_publickeys))

		}

		//  fmt.Printf("len of publickeylist  %d \n", len(publickeylist))

		//  revealing r will disclose the amount and the sender and receiver and separate anonymous ring members
		// calculate r deterministically, so its different every transaction, in emergency it can be given to other, and still will not allows key attacks
		rinputs := append([]byte{}, hash[:]...)
		for i := range publickeylist {
			rinputs = append(rinputs, publickeylist[i].EncodeCompressed()...)
		}
		rencrypted := new(bn256.G1).ScalarMult(crypto.HashToPoint(crypto.HashtoNumber(append([]byte(crypto.PROTOCOL_CONSTANT), rinputs...))), sender_secret)
		r := crypto.ReducedHash(rencrypted.EncodeCompressed())

		//r := crypto.RandomScalarFixed()

		//fmt.Printf("r %s\n", r.Text(16))

		var payload transaction_zether.TransactionZetherPayload

		payload.Token = transfers[t].Token
		payload.BurnValue = transfers[t].Burn

		fees := uint64(0)
		value := transfers[t].Amount
		burn_value := transfers[t].Burn

		for i := range publickeylist { // setup commitments
			var x bn256.G1
			switch {
			case i == witness_index[0]:
				x.ScalarMult(crypto.G, new(big.Int).SetInt64(0-int64(value)-int64(fees)-int64(burn_value))) // decrease senders balance
				//fmt.Printf("sender %s \n", x.String())
			case i == witness_index[1]:
				x.ScalarMult(crypto.G, new(big.Int).SetInt64(int64(value))) // increase receiver's balance
				//fmt.Printf("receiver %s \n", x.String())

				// lets encrypt the payment id, it's simple, we XOR the paymentID
				blinder := new(bn256.G1).ScalarMult(publickeylist[i], r)

				// we must obfuscate it for non-client call
				if len(publickeylist) >= 512 {
					return nil, errors.New("currently we donot support ring size >= 512")
				}

				payload.ExtraType = transaction_zether.ENCRYPTED_DEFAULT_PAYLOAD_CBOR

				var dataFinal []byte
				if dataFinal, err = transfer.Data.getData(); err != nil {
					return
				}
				if len(dataFinal) > transaction_zether.PAYLOAD0_LIMIT {
					return nil, errors.New("Data final exceeds")
				}
				dataFinal = append(dataFinal, make([]byte, transaction_zether.PAYLOAD0_LIMIT-len(dataFinal))...)

				payload.ExtraData = append([]byte{byte(uint(witness_index[0]))}, dataFinal...)

				//fmt.Printf("%d packed rpc payload %d %x\n ", t, len(data), data)
				// make sure used data encryption is optional, just in case we would like to play together with ring members
				if err = crypto.EncryptDecryptUserData(blinder, payload.ExtraData); err != nil {
					return
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
			var ebalance *crypto.ElGamal

			switch {
			case i == witness_index[0]:
				if ebalance, err = new(crypto.ElGamal).Deserialize(emap[string(transfers[t].Token)][sender.String()]); err != nil {
					return
				}
			case i == witness_index[1]:
				if ebalance, err = new(crypto.ElGamal).Deserialize(emap[string(transfers[t].Token)][receiver.String()]); err != nil {
					return
				}
				//fmt.Printf("receiver %s \n", x.String())
			default:
				//x.ScalarMult(crypto.G, new(big.Int).SetInt64(0))
				// panic("anon ring currently not supported")
				ebalance = ebalances_list[i]
			}

			var ll, rr bn256.G1
			//ebalance := b.balances[publickeylist[i].String()] // note these are taken from the chain live

			ll.Add(ebalance.Left, C[i])
			CLn = append(CLn, &ll)
			//  fmt.Printf("%d CLnG %x\n", i,CLn[i].EncodeCompressed())

			rr.Add(ebalance.Right, &D)
			CRn = append(CRn, &rr)
			//  fmt.Printf("%d CRnG %x\n",i, CRn[i].EncodeCompressed())

		}

		// decode balance now
		var pt *crypto.ElGamal
		if pt, err = new(crypto.ElGamal).Deserialize(emap[string(transfers[t].Token)][sender.String()]); err != nil {
			return
		}
		balance := senderKey.DecodeBalance(pt, transfer.FromBalanceDecoded)

		//fmt.Printf("t %d scid %s  balance %d\n", t, transfers[t].SCID, balance)

		// time for bullets-sigma
		statement := GenerateStatement(CLn, CRn, publickeylist, C, &D, fees) // generate statement
		statement.Roothash = make([]byte, 32)
		copy(statement.Roothash[:], hash[:])

		statement.RingSize = uint64(len(publickeylist))

		witness := GenerateWitness(sender_secret, r, value, balance-value-fees-burn_value, witness_index)

		witness_list = append(witness_list, witness)

		// this goes to proof.u

		//Print(statement, witness)
		payload.Statement = &statement

		txBase.Payloads = append(txBase.Payloads, &payload)

		// get ready for another round by internal processing of state
		for i := range publickeylist {

			var balance *crypto.ElGamal
			if balance, err = new(crypto.ElGamal).Deserialize(emap[string(transfers[t].Token)][publickeylist[i].String()]); err != nil {
				return
			}
			echanges := crypto.ConstructElGamal(statement.C[i], statement.D)

			balance = balance.Add(echanges)                                                   // homomorphic addition of changes
			emap[string(transfers[t].Token)][publickeylist[i].String()] = balance.Serialize() // reserialize and store
		}

	}

	senderKey := &addresses.PrivateKey{Key: transfers[0].From}
	sender_secret := new(crypto.BNRed).SetBytes(senderKey.Key).BigInt()

	u := new(bn256.G1).ScalarMult(crypto.HeightToPoint(height), sender_secret)                          // this should be moved to generate proof
	u1 := new(bn256.G1).ScalarMult(crypto.HeightToPoint(height+crypto.BLOCK_BATCH_SIZE), sender_secret) // this should be moved to generate proof

	for t := range transfers {
		if txBase.Payloads[t].Proof, err = crypto.GenerateProof(txBase.Payloads[t].Statement, &witness_list[t], u, u1, height, tx.GetHashSigningManually(), txBase.Payloads[t].BurnValue); err != nil {
			return
		}
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	statusCallback("Transaction Bloomed")

	if err = tx.Verify(); err != nil {
		return
	}
	statusCallback("Transaction Verified")

	return tx, nil
}

// generate statement
func GenerateStatement(CLn, CRn, publickeylist, C []*bn256.G1, D *bn256.G1, fees uint64) crypto.Statement {
	return crypto.Statement{CLn: CLn, CRn: CRn, Publickeylist: publickeylist, C: C, D: D, Fees: fees}
}

// generate witness
func GenerateWitness(secretkey, r *big.Int, TransferAmount, Balance uint64, index []int) crypto.Witness {
	return crypto.Witness{SecretKey: secretkey, R: r, TransferAmount: TransferAmount, Balance: Balance, Index: index}
}
