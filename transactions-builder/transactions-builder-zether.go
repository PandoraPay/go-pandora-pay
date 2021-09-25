package transactions_builder

import (
	"encoding/binary"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/blockchain/data/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	advanced_connection_types "pandora-pay/network/websocks/connection/advanced-connection-types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/transactions-builder/wizard"
)

func (builder *TransactionsBuilder) CreateZetherTx_Float(from []string, tokensUsed [][]byte, amounts []float64, dsts []string, burn []float64, ringMembers [][]string, data []*wizard.TransactionsWizardData, fees []*TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	amountsFinal := make([]uint64, len(amounts))
	burnFinal := make([]uint64, len(burn))
	finalFees := make([]*wizard.TransactionsWizardFee, len(fees))

	statusCallback("Converting Floats to Numbers")

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)

		for i := range amounts {
			if err != nil {
				return
			}

			var tok *token.Token
			if tok, err = toks.GetToken(tokensUsed[i]); err != nil {
				return
			}
			if tok == nil {
				return errors.New("Token was not found")
			}

			if amountsFinal[i], err = tok.ConvertToUnits(amounts[i]); err != nil {
				return
			}
			if burnFinal[i], err = tok.ConvertToUnits(burn[i]); err != nil {
				return
			}
			if finalFees[i], err = fees[i].convertToWizardFee(tok); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateZetherTx(from, tokensUsed, amountsFinal, dsts, burnFinal, ringMembers, data, finalFees, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateZetherTx(from []string, tokensUsed [][]byte, amounts []uint64, dsts []string, burn []uint64, ringMembers [][]string, data []*wizard.TransactionsWizardData, fees []*wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	if len(from) != len(tokensUsed) || len(tokensUsed) != len(amounts) || len(amounts) != len(dsts) || len(dsts) != len(burn) || len(burn) != len(data) || len(data) != len(fees) {
		return nil, errors.New("Length of from and transfers are not matching")
	}

	fromWalletAddresses, err := builder.getWalletAddresses(from)
	if err != nil {
		return nil, err
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	var chainHeight uint64
	var chainHash []byte

	transfers := make([]*wizard.ZetherTransfer, len(from))

	emap := make(map[string]map[string][]byte) //initialize all maps
	rings := make([][]*bn256.G1, len(from))

	publicKeyIndexes := make(map[string]*wizard.ZetherPublicKeyIndex)

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accsCollection := accounts.NewAccountsCollection(reader)
		regs := registrations.NewRegistrations(reader)

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		chainHash = helpers.CloneBytes(reader.Get("chainHash"))

		for _, token := range tokensUsed {
			if emap[string(token)] == nil {
				emap[string(token)] = map[string][]byte{}
			}
		}

		for i, fromWalletAddress := range fromWalletAddresses {

			var accs *accounts.Accounts
			if accs, err = accsCollection.GetMap(tokensUsed[i]); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccount(fromWalletAddress.PublicKey); err != nil {
				return
			}

			var fromBalanceDecoded uint64
			if fromBalanceDecoded, err = builder.wallet.DecodeBalanceByPublicKey(fromWalletAddress.PublicKey, acc.GetBalance(), acc.Token, false); err != nil {
				return
			}

			transfers[i] = &wizard.ZetherTransfer{
				Token:              tokensUsed[i],
				From:               fromWalletAddress.PrivateKey.Key[:],
				FromBalanceDecoded: fromBalanceDecoded,
				Destination:        dsts[i],
				Amount:             amounts[i],
				Burn:               burn[i],
				Data:               data[i],
			}

			if fromBalanceDecoded < amounts[i] {
				return errors.New("Not enough funds")
			}

			var ring []*bn256.G1

			addPoint := func(address string) (err error) {
				var addr *addresses.Address
				var p *crypto.Point
				if addr, err = addresses.DecodeAddr(address); err != nil {
					return
				}
				if p, err = addr.GetPoint(); err != nil {
					return
				}

				if acc, err = accs.GetAccount(addr.PublicKey); err != nil {
					return
				}

				var ebalance *crypto.ElGamal
				if acc != nil {
					ebalance = acc.GetBalance()
				} else {
					ebalance = crypto.ConstructElGamal(p.G1(), crypto.ElGamal_BASE_G)
				}
				emap[string(tokensUsed[i])][p.G1().String()] = ebalance.Serialize()

				ring = append(ring, p.G1())

				var isReg bool
				if isReg, err = regs.Exists(string(addr.PublicKey)); err != nil {
					return
				}

				publicKeyIndex := &wizard.ZetherPublicKeyIndex{}
				publicKeyIndexes[string(addr.PublicKey)] = publicKeyIndex

				publicKeyIndex.Registered = isReg
				if isReg {
					if publicKeyIndex.RegisteredIndex, err = regs.GetIndexByKey(string(addr.PublicKey)); err != nil {
						return
					}
				} else {
					publicKeyIndex.RegistrationSignature = addr.Registration
				}

				return
			}

			if err = addPoint(fromWalletAddress.AddressEncoded); err != nil {
				return
			}
			if err = addPoint(dsts[i]); err != nil {
				return
			}

			for _, ringMember := range ringMembers[i] {
				if err = addPoint(ringMember); err != nil {
					return
				}
			}

			rings[i] = ring
		}
		statusCallback("Wallet Addresses Found")

		return
	}); err != nil {
		return nil, err
	}

	statusCallback("Balances checked")

	if tx, err = wizard.CreateZetherTx(transfers, emap, rings, chainHeight, chainHash, publicKeyIndexes, fees, statusCallback); err != nil {
		gui.GUI.Error("Error creating Tx: ", err)
		return nil, err
	}

	statusCallback("Transaction Created")

	if propagateTx {
		if err := builder.mempool.AddTxToMemPool(tx, chainHeight, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
			return nil, err
		}
	}

	return tx, nil
}
