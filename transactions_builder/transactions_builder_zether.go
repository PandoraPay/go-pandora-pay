package transactions_builder

import (
	"context"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/wallet/wallet_address"
)

func (builder *TransactionsBuilder) getRandomAccount(accs *accounts.Accounts) (addr *addresses.Address, err error) {

	var acc *account.Account

	if acc, err = accs.GetRandomAccount(); err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, errors.New("Error getting any random account")
	}

	if addr, err = addresses.CreateAddr(acc.PublicKey, nil, 0, nil); err != nil {
		return nil, err
	}

	return
}

func (builder *TransactionsBuilder) createZetherRing(from string, dst *string, assetId []byte, ringConfiguration *ZetherRingConfiguration, dataStorage *data_storage.DataStorage) ([]string, error) {

	var addr *addresses.Address
	var err error

	if ringConfiguration.RingSize == -1 {
		pow := rand.Intn(4) + 4
		ringConfiguration.RingSize = int(math.Pow(2, float64(pow)))
	}
	if ringConfiguration.NewAccounts == -1 {
		ringConfiguration.NewAccounts = rand.Intn(ringConfiguration.RingSize / 5)
	}

	if ringConfiguration.RingSize < 0 {
		return nil, errors.New("number is negative")
	}
	if !crypto.IsPowerOf2(ringConfiguration.RingSize) {
		return nil, errors.New("ring size is not a power of 2")
	}
	if ringConfiguration.NewAccounts < 0 || ringConfiguration.NewAccounts > ringConfiguration.RingSize-2 {
		return nil, errors.New("New accounts needs to be in the interval [0, ringSize-2] ")
	}

	var accs *accounts.Accounts

	if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
		return nil, err
	}

	alreadyUsed := make(map[string]bool)

	if addr, err = addresses.DecodeAddr(from); err != nil {
		return nil, err
	}
	alreadyUsed[string(addr.PublicKey)] = true

	if *dst == "" {
		if addr, err = builder.getRandomAccount(accs); err != nil {
			return nil, err
		}
		*dst = addr.EncodeAddr()
	}

	if addr, err = addresses.DecodeAddr(*dst); err != nil {
		return nil, err
	}
	alreadyUsed[string(addr.PublicKey)] = true

	rings := make([]string, ringConfiguration.RingSize-2)

	if globals.Arguments["--new-devnet"] == true && accs.Count < 80000 {
		ringConfiguration.NewAccounts = ringConfiguration.RingSize - 2
	}

	for i := 0; i < ringConfiguration.RingSize-2; i++ {

		if i < ringConfiguration.NewAccounts || accs.Count-2+uint64(ringConfiguration.NewAccounts) <= uint64(i) {
			priv := addresses.GenerateNewPrivateKey()
			if addr, err = priv.GenerateAddress(true, 0, nil); err != nil {
				return nil, err
			}
		} else {
			if addr, err = builder.getRandomAccount(accs); err != nil {
				return nil, err
			}
		}

		if alreadyUsed[string(addr.PublicKey)] {
			i--
			continue
		}
		alreadyUsed[string(addr.PublicKey)] = true
		rings[i] = addr.EncodeAddr()
	}

	return rings, nil
}

func (builder *TransactionsBuilder) prebuild(extraPayloads []wizard.WizardZetherPayloadExtra, from []string, dstsAsts [][]byte, amounts []uint64, dsts []string, burns []uint64, ringsConfiguration []*ZetherRingConfiguration, data []*wizard.TransactionsWizardData, fees []*wizard.TransactionsWizardFee, ctx context.Context, statusCallback func(string)) ([]*wizard.WizardZetherTransfer, map[string]map[string][]byte, [][]*bn256.G1, map[string]*wizard.WizardZetherPublicKeyIndex, uint64, []byte, error) {

	if len(from) != len(dstsAsts) || len(dstsAsts) != len(amounts) || len(amounts) != len(dsts) || len(dsts) != len(burns) || len(burns) != len(data) || len(data) != len(fees) {
		return nil, nil, nil, nil, 0, nil, errors.New("Length of from and transfers are not matching")
	}

	fromPrivateKeys := make([]*addresses.PrivateKey, len(from))
	fromWalletAddresses := make([]*wallet_address.WalletAddress, len(from))

	for t := range from {
		if from[t] == "" {

			fromPrivateKeys[t] = addresses.GenerateNewPrivateKey()
			addr, err := fromPrivateKeys[t].GenerateAddress(true, 0, nil)
			if err != nil {
				return nil, nil, nil, nil, 0, nil, err
			}
			from[t] = addr.EncodeAddr()

		} else {

			addr, err := builder.wallet.GetWalletAddressByEncodedAddress(from[t])
			if err != nil {
				return nil, nil, nil, nil, 0, nil, err
			}

			if addr.PrivateKey == nil {
				return nil, nil, nil, nil, 0, nil, errors.New("Can't be used for transactions as the private key is missing")
			}

			fromPrivateKeys[t] = &addresses.PrivateKey{Key: addr.PrivateKey.Key[:]}
			from[t] = addr.AddressRegistrationEncoded
			fromWalletAddresses[t] = addr
		}

	}

	ringMembers := make([][]string, len(from))

	var chainHeight uint64
	var chainHash []byte

	transfers := make([]*wizard.WizardZetherTransfer, len(from))
	emap := wizard.InitializeEmap(dstsAsts)
	rings := make([][]*bn256.G1, len(from))
	publicKeyIndexes := make(map[string]*wizard.WizardZetherPublicKeyIndex)

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.CreateDataStorage(reader)

		for t := range from {
			if ringMembers[t], err = builder.createZetherRing(from[t], &dsts[t], dstsAsts[t], ringsConfiguration[t], dataStorage); err != nil {
				return
			}
		}

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		chainHash = helpers.CloneBytes(reader.Get("chainHash"))

		for t, ast := range dstsAsts {

			var accs *accounts.Accounts
			if accs, err = dataStorage.AccsCollection.GetMap(ast); err != nil {
				return
			}

			transfers[t] = &wizard.WizardZetherTransfer{
				Asset:        ast,
				From:         fromPrivateKeys[t].Key[:],
				Destination:  dsts[t],
				Amount:       amounts[t],
				Burn:         burns[t],
				Data:         data[t],
				PayloadExtra: extraPayloads[t],
			}

			var ring []*bn256.G1
			uniqueMap := make(map[string]bool)

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

				if emap[string(ast)][p.G1().String()] == nil {

					var acc *account.Account
					if acc, err = accs.GetAccount(addr.PublicKey); err != nil {
						return
					}

					var balance []byte
					if acc != nil {
						balance = helpers.CloneBytes(acc.Balance.Amount.Serialize())
					}

					if balance, err = builder.mempool.GetZetherBalance(addr.PublicKey, balance); err != nil {
						return
					}

					if balance == nil {
						var acckey crypto.Point
						if err = acckey.DecodeCompressed(addr.PublicKey); err != nil {
							return
						}
						balance = crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G).Serialize()
					}

					if from[t] == address { //sender

						if fromWalletAddresses[t] == nil {
							transfers[t].FromBalanceDecoded = transfers[t].Amount
						} else {
							var balancePoint *crypto.ElGamal
							if balancePoint, err = new(crypto.ElGamal).Deserialize(balance); err != nil {
								return
							}

							if transfers[t].FromBalanceDecoded, err = builder.wallet.DecodeBalanceByPublicKey(fromWalletAddresses[t].PublicKey, balancePoint, ast, true, true, ctx, statusCallback); err != nil {
								return
							}

						}
						if transfers[t].FromBalanceDecoded == 0 {
							return errors.New("You have no funds")
						}
						if transfers[t].FromBalanceDecoded < amounts[t] {
							return errors.New("Not enough funds")
						}
					}

					emap[string(ast)][p.G1().String()] = balance
				}
				ring = append(ring, p.G1())

				if publicKeyIndexes[string(addr.PublicKey)] == nil {
					var reg *registration.Registration
					if reg, err = dataStorage.Regs.GetRegistration(addr.PublicKey); err != nil {
						return
					}

					publicKeyIndex := &wizard.WizardZetherPublicKeyIndex{}
					publicKeyIndexes[string(addr.PublicKey)] = publicKeyIndex

					if reg != nil {
						publicKeyIndex.Registered = true
						publicKeyIndex.RegisteredIndex = reg.Index
					} else {
						publicKeyIndex.RegistrationSignature = addr.Registration
					}
				}

				return
			}

			if err = addPoint(from[t]); err != nil {
				return
			}
			if err = addPoint(dsts[t]); err != nil {
				return
			}
			for _, ringMember := range ringMembers[t] {
				if err = addPoint(ringMember); err != nil {
					return
				}
			}

			rings[t] = ring
		}
		statusCallback("Wallet Addresses Found")

		return
	}); err != nil {
		return nil, nil, nil, nil, 0, nil, err
	}
	statusCallback("Balances checked")

	return transfers, emap, rings, publicKeyIndexes, chainHeight, chainHash, nil
}

func (builder *TransactionsBuilder) CreateZetherTx(extraPayloads []wizard.WizardZetherPayloadExtra, from []string, asts [][]byte, amounts []uint64, dsts []string, burns []uint64, ringsConfiguration []*ZetherRingConfiguration, data []*wizard.TransactionsWizardData, fees []*wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast bool, validateTx bool, ctx context.Context, statusCallback func(string)) (*transaction.Transaction, error) {

	builder.lock.Lock()
	defer builder.lock.Unlock()

	transfers, emap, ringMembers, publicKeyIndexes, chainHeight, chainHash, err := builder.prebuild(extraPayloads, from, asts, amounts, dsts, burns, ringsConfiguration, data, fees, ctx, statusCallback)
	if err != nil {
		return nil, err
	}

	var tx *transaction.Transaction
	if tx, err = wizard.CreateZetherTx(transfers, emap, ringMembers, chainHeight, chainHash, publicKeyIndexes, fees, validateTx, ctx, statusCallback); err != nil {
		return nil, err
	}

	if propagateTx {
		if err := builder.mempool.AddTxToMemPool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
			return nil, err
		}
	}

	return tx, nil
}
