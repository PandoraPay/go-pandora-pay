package txs_builder

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/txs_builder/wizard"
	"pandora-pay/wallet/wallet_address"
)

func (builder *TxsBuilder) getRandomAccount(accs *accounts.Accounts) (addr *addresses.Address, err error) {

	var acc *account.Account

	if acc, err = accs.GetRandomAccount(); err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, errors.New("Error getting any random account")
	}

	if addr, err = addresses.CreateAddr(acc.PublicKey, nil, nil, 0, nil); err != nil {
		return nil, err
	}

	return
}

func (builder *TxsBuilder) createZetherRing(sender string, recipient *string, assetId []byte, ringConfiguration *ZetherRingConfiguration, dataStorage *data_storage.DataStorage) ([]string, error) {

	var addr *addresses.Address
	var err error

	if ringConfiguration.RingSize == -1 {
		probability := rand.Intn(1000)
		if probability < 400 {
			ringConfiguration.RingSize = 32
		} else if probability < 600 {
			ringConfiguration.RingSize = 64
		} else if probability < 800 {
			ringConfiguration.RingSize = 128
		} else {
			ringConfiguration.RingSize = 256
		}
	}
	if ringConfiguration.NewAccounts == -1 {
		probability := rand.Intn(1000)
		if probability < 800 {
			ringConfiguration.NewAccounts = 0
		} else if probability < 900 {
			ringConfiguration.NewAccounts = 1
		} else {
			ringConfiguration.NewAccounts = 2
		}
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

	if addr, err = addresses.DecodeAddr(sender); err != nil {
		return nil, err
	}
	alreadyUsed[string(addr.PublicKey)] = true

	if *recipient == "" {
		if addr, err = builder.getRandomAccount(accs); err != nil {
			return nil, err
		}
		*recipient = addr.EncodeAddr()
	}

	if addr, err = addresses.DecodeAddr(*recipient); err != nil {
		return nil, err
	}
	alreadyUsed[string(addr.PublicKey)] = true

	ring := make([]string, 0)

	if ringConfiguration.IncludeMembers != nil {
		for _, member := range ringConfiguration.IncludeMembers {
			if addr, err = addresses.DecodeAddr(member); err != nil {
				return nil, err
			}
			if alreadyUsed[string(addr.PublicKey)] {
				continue
			}
			alreadyUsed[string(addr.PublicKey)] = true
			ring = append(ring, addr.EncodeAddr())
		}
	}

	if globals.Arguments["--new-devnet"] == true && accs.Count < 80000 {
		ringConfiguration.NewAccounts = ringConfiguration.RingSize - 2
	}

	for i := 0; i < ringConfiguration.NewAccounts && len(ring) < ringConfiguration.RingSize-2; i++ {
		priv := addresses.GenerateNewPrivateKey()
		if addr, err = priv.GenerateAddress(true, nil, 0, nil); err != nil {
			return nil, err
		}
		if alreadyUsed[string(addr.PublicKey)] {
			i--
			continue
		}
		alreadyUsed[string(addr.PublicKey)] = true
		ring = append(ring, addr.EncodeAddr())
	}

	for len(ring) < ringConfiguration.RingSize-2 {

		if accs.Count-2+uint64(ringConfiguration.NewAccounts) <= uint64(ringConfiguration.RingSize) {
			priv := addresses.GenerateNewPrivateKey()
			if addr, err = priv.GenerateAddress(true, nil, 0, nil); err != nil {
				return nil, err
			}
		} else {
			if addr, err = builder.getRandomAccount(accs); err != nil {
				return nil, err
			}
		}

		if alreadyUsed[string(addr.PublicKey)] {
			continue
		}
		alreadyUsed[string(addr.PublicKey)] = true
		ring = append(ring, addr.EncodeAddr())
	}

	return ring, nil
}

func (builder *TxsBuilder) prebuild(txData *TxBuilderCreateZetherTxData, ctx context.Context, statusCallback func(string)) ([]*wizard.WizardZetherTransfer, map[string]map[string][]byte, [][]*bn256.G1, map[string]*wizard.WizardZetherPublicKeyIndex, uint64, []byte, error) {

	sendersPrivateKeys := make([]*addresses.PrivateKey, len(txData.Payloads))
	sendersWalletAddresses := make([]*wallet_address.WalletAddress, len(txData.Payloads))
	sendAssets := make([][]byte, len(txData.Payloads))

	for t, payload := range txData.Payloads {

		if payload.Asset == nil {
			payload.Asset = config_coins.NATIVE_ASSET_FULL
		}
		if payload.Data == nil {
			payload.Data = &wizard.WizardTransactionData{[]byte{}, false}
		}
		if payload.RingConfiguration == nil {
			payload.RingConfiguration = &ZetherRingConfiguration{-1, -1, nil}
		}
		if payload.Fee == nil {
			payload.Fee = &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}
		}

		sendAssets[t] = payload.Asset
		if payload.Sender == "" {

			sendersPrivateKeys[t] = addresses.GenerateNewPrivateKey()
			addr, err := sendersPrivateKeys[t].GenerateAddress(true, nil, 0, nil)
			if err != nil {
				return nil, nil, nil, nil, 0, nil, err
			}
			payload.Sender = addr.EncodeAddr()

		} else {

			addr, err := builder.wallet.GetWalletAddressByEncodedAddress(payload.Sender, true)
			if err != nil {
				return nil, nil, nil, nil, 0, nil, err
			}

			if addr.PrivateKey == nil {
				return nil, nil, nil, nil, 0, nil, errors.New("Can't be used for transactions as the private key is missing")
			}

			sendersPrivateKeys[t] = &addresses.PrivateKey{Key: addr.PrivateKey.Key[:]}
			payload.Sender = addr.AddressRegistrationEncoded
			sendersWalletAddresses[t] = addr
		}

	}

	ringMembers := make([][]string, len(txData.Payloads))

	var chainHeight uint64
	var chainHash []byte

	transfers := make([]*wizard.WizardZetherTransfer, len(txData.Payloads))
	emap := wizard.InitializeEmap(sendAssets)
	rings := make([][]*bn256.G1, len(txData.Payloads))
	publicKeyIndexes := make(map[string]*wizard.WizardZetherPublicKeyIndex)

	sendersEncryptedBalances := make([][]byte, len(txData.Payloads))

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		for t, payload := range txData.Payloads {
			if ringMembers[t], err = builder.createZetherRing(payload.Sender, &payload.Recipient, payload.Asset, payload.RingConfiguration, dataStorage); err != nil {
				return
			}
		}

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		chainHash = reader.Get("chainHash")

		for t, payload := range txData.Payloads {

			var accs *accounts.Accounts
			if accs, err = dataStorage.AccsCollection.GetMap(payload.Asset); err != nil {
				return
			}

			if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) && payload.Fee.Auto {
				var assetFeeLiquidity *asset_fee_liquidity.AssetFeeLiquidity
				if assetFeeLiquidity, err = dataStorage.GetAssetFeeLiquidityTop(payload.Asset, chainHeight); err != nil {
					return
				}
				if assetFeeLiquidity == nil {
					return errors.New("There is no Asset Fee Liquidity for this asset")
				}
				payload.Fee.Rate = assetFeeLiquidity.Rate
				payload.Fee.LeadingZeros = assetFeeLiquidity.LeadingZeros
			}

			transfers[t] = &wizard.WizardZetherTransfer{
				Asset:           payload.Asset,
				Sender:          sendersPrivateKeys[t].Key[:],
				Recipient:       payload.Recipient,
				Amount:          payload.Amount,
				Burn:            payload.Burn,
				Data:            payload.Data,
				FeeRate:         payload.Fee.Rate,
				FeeLeadingZeros: payload.Fee.LeadingZeros,
				PayloadExtra:    payload.Extra,
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

				if emap[string(payload.Asset)][p.G1().String()] == nil {

					var acc *account.Account
					if acc, err = accs.GetAccount(addr.PublicKey, chainHeight); err != nil {
						return
					}

					var balance []byte = nil
					if acc != nil {
						balance = acc.Balance.Amount.Serialize()
					}

					if balance, err = builder.mempool.GetZetherBalance(addr.PublicKey, balance, payload.Asset); err != nil {
						return
					}

					if payload.Sender == address { //sender
						sendersEncryptedBalances[t] = balance
					}

					emap[string(payload.Asset)][p.G1().String()] = balance
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

			if err = addPoint(payload.Sender); err != nil {
				return
			}
			if err = addPoint(payload.Recipient); err != nil {
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

	for t := range transfers {
		if sendersWalletAddresses[t] == nil {
			transfers[t].SenderDecryptedBalance = transfers[t].Amount
		} else {

			var err error
			if transfers[t].SenderDecryptedBalance, err = builder.wallet.DecryptBalanceByPublicKey(sendersWalletAddresses[t].PublicKey, sendersEncryptedBalances[t], transfers[t].Asset, false, 0, true, true, ctx, statusCallback); err != nil {
				return nil, nil, nil, nil, 0, nil, err
			}

		}
		if transfers[t].SenderDecryptedBalance == 0 {
			return nil, nil, nil, nil, 0, nil, errors.New("You have no funds")
		}
		if transfers[t].SenderDecryptedBalance < txData.Payloads[t].Amount {
			return nil, nil, nil, nil, 0, nil, errors.New("Not enough funds")
		}
	}

	statusCallback("Balances decoded")

	return transfers, emap, rings, publicKeyIndexes, chainHeight, chainHash, nil
}

func (builder *TxsBuilder) CreateZetherTx(txData *TxBuilderCreateZetherTxData, propagateTx, awaitAnswer, awaitBroadcast bool, validateTx bool, ctx context.Context, statusCallback func(string)) (*transaction.Transaction, error) {

	builder.lock.Lock()
	defer builder.lock.Unlock()

	transfers, emap, ringMembers, publicKeyIndexes, chainHeight, chainHash, err := builder.prebuild(txData, ctx, statusCallback)
	if err != nil {
		return nil, err
	}

	feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		feesFinal[t] = payload.Fee.WizardTransactionFee
	}

	var tx *transaction.Transaction
	if tx, err = wizard.CreateZetherTx(transfers, emap, ringMembers, chainHeight, chainHash, publicKeyIndexes, feesFinal, validateTx, ctx, statusCallback); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMempool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL, ctx); err != nil {
			return nil, err
		}
	}

	return tx, nil
}
