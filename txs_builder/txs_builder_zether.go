package txs_builder

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_reward"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
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

func (builder *TxsBuilder) presetZetherRing(ringConfiguration *ZetherRingConfiguration) error {

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
	if ringConfiguration.RecipientRingType.NewAccounts == -1 {
		probability := rand.Intn(1000)
		if probability < 800 {
			ringConfiguration.RecipientRingType.NewAccounts = 0
		} else if probability < 900 {
			ringConfiguration.RecipientRingType.NewAccounts = 1
		} else {
			ringConfiguration.RecipientRingType.NewAccounts = 2
		}
	}

	if ringConfiguration.RingSize < 0 {
		return errors.New("number is negative")
	}
	if !crypto.IsPowerOf2(ringConfiguration.RingSize) {
		return errors.New("ring size is not a power of 2")
	}
	if ringConfiguration.RecipientRingType.NewAccounts < 0 || ringConfiguration.RecipientRingType.NewAccounts > ringConfiguration.RingSize/2-1 {
		return errors.New("New accounts needs to be in the interval [0, ringSize-2] ")
	}

	return nil
}

func (builder *TxsBuilder) createZetherRing(sender, receiver *string, assetId []byte, ringConfiguration *ZetherRingConfiguration, dataStorage *data_storage.DataStorage) ([]string, []string, error) {

	var addr *addresses.Address
	var err error

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
		return nil, nil, err
	}

	alreadyUsed := make(map[string]bool)

	setAddress := func(address *string) (err error) {
		if *address == "" {
			if accs.Count == uint64(len(alreadyUsed)) {
				return errors.New("Accounts have only member. Impossible to get random recipient")
			}
			for {
				if addr, err = builder.getRandomAccount(accs); err != nil {
					return err
				}
				if alreadyUsed[string(addr.PublicKey)] {
					continue
				}
				*address = addr.EncodeAddr()
				break
			}
		} else {
			if addr, err = addresses.DecodeAddr(*address); err != nil {
				return err
			}
		}
		if alreadyUsed[string(addr.PublicKey)] {
			return errors.New("Address was used before")
		}
		alreadyUsed[string(addr.PublicKey)] = true
		return nil
	}

	includeMembers := func(ring *[]string, includeMembers []string) (err error) {
		if includeMembers != nil {
			for _, member := range includeMembers {
				if addr, err = addresses.DecodeAddr(member); err != nil {
					return err
				}
				if alreadyUsed[string(addr.PublicKey)] {
					continue
				}
				alreadyUsed[string(addr.PublicKey)] = true
				*ring = append(*ring, addr.EncodeAddr())
			}
		}
		return nil
	}

	newAccounts := func(ring *[]string, newAccounts int) (err error) {
		for i := 0; i < newAccounts && len(*ring) < ringConfiguration.RingSize/2-1; i++ {
			priv := addresses.GenerateNewPrivateKey()
			if addr, err = priv.GenerateAddress(true, nil, 0, nil); err != nil {
				return
			}
			if alreadyUsed[string(addr.PublicKey)] {
				i--
				continue
			}
			alreadyUsed[string(addr.PublicKey)] = true
			*ring = append(*ring, addr.EncodeAddr())
		}
		return
	}

	newRandomAccounts := func(ring *[]string) (err error) {

		for len(*ring) < ringConfiguration.RingSize/2-1 {

			if accs.Count <= uint64(len(alreadyUsed)) {
				priv := addresses.GenerateNewPrivateKey()
				if addr, err = priv.GenerateAddress(true, nil, 0, nil); err != nil {
					return
				}
			} else {
				if addr, err = builder.getRandomAccount(accs); err != nil {
					return
				}
			}

			if alreadyUsed[string(addr.PublicKey)] {
				continue
			}
			alreadyUsed[string(addr.PublicKey)] = true
			*ring = append(*ring, addr.EncodeAddr())
		}

		return
	}

	if err = setAddress(sender); err != nil {
		return nil, nil, err
	}
	if err = setAddress(receiver); err != nil {
		return nil, nil, err
	}

	senderRing := make([]string, 0)
	recipientRing := make([]string, 0)

	if err = includeMembers(&senderRing, ringConfiguration.SenderRingType.IncludeMembers); err != nil {
		return nil, nil, err
	}
	if err = includeMembers(&recipientRing, ringConfiguration.RecipientRingType.IncludeMembers); err != nil {
		return nil, nil, err
	}

	if err = newAccounts(&senderRing, ringConfiguration.SenderRingType.NewAccounts); err != nil {
		return nil, nil, err
	}
	if err = newAccounts(&recipientRing, ringConfiguration.RecipientRingType.NewAccounts); err != nil {
		return nil, nil, err
	}

	if err = newRandomAccounts(&senderRing); err != nil {
		return nil, nil, err
	}
	if err = newRandomAccounts(&recipientRing); err != nil {
		return nil, nil, err
	}

	return senderRing, recipientRing, err
}

func (builder *TxsBuilder) prebuild(txData *TxBuilderCreateZetherTxData, ctx context.Context, statusCallback func(string)) ([]*wizard.WizardZetherTransfer, map[string]map[string][]byte, map[string]bool, [][]*bn256.G1, map[string]*wizard.WizardZetherPublicKeyIndex, uint64, []byte, error) {

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
			payload.RingConfiguration = &ZetherRingConfiguration{-1, &ZetherSenderRingType{false, nil, 0}, &ZetherRecipientRingType{false, nil, 0}}
		}
		if payload.Fee == nil {
			payload.Fee = &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}
		}

		sendAssets[t] = payload.Asset
		if payload.Sender == "" {

			sendersPrivateKeys[t] = addresses.GenerateNewPrivateKey()
			addr, err := sendersPrivateKeys[t].GenerateAddress(true, nil, 0, nil)
			if err != nil {
				return nil, nil, nil, nil, nil, 0, nil, err
			}
			payload.Sender = addr.EncodeAddr()

		} else {

			addr, err := builder.wallet.GetWalletAddressByEncodedAddress(payload.Sender, true)
			if err != nil {
				return nil, nil, nil, nil, nil, 0, nil, err
			}

			if addr.PrivateKey == nil {
				return nil, nil, nil, nil, nil, 0, nil, errors.New("Can't be used for transactions as the private key is missing")
			}

			sendersPrivateKeys[t] = &addresses.PrivateKey{Key: addr.PrivateKey.Key[:]}
			payload.Sender = addr.AddressRegistrationEncoded
			sendersWalletAddresses[t] = addr
		}

	}

	senderRingMembers := make([][]string, len(txData.Payloads))
	recipientRingMembers := make([][]string, len(txData.Payloads))

	var chainHeight uint64
	var chainKernelHash []byte

	transfers := make([]*wizard.WizardZetherTransfer, len(txData.Payloads))
	emap := wizard.InitializeEmap(sendAssets)
	hasRollovers := make(map[string]bool)

	rings := make([][]*bn256.G1, len(txData.Payloads))
	publicKeyIndexes := make(map[string]*wizard.WizardZetherPublicKeyIndex)

	sendersEncryptedBalances := make([][]byte, len(txData.Payloads))

	for _, payload := range txData.Payloads {
		if err := builder.presetZetherRing(payload.RingConfiguration); err != nil {
			return nil, nil, nil, nil, nil, 0, nil, err
		}
	}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		for t, payload := range txData.Payloads {

			if payload.Extra != nil {
				switch payload.Extra.(type) {
				case *wizard.WizardZetherPayloadExtraStakingReward:

					recipientRingMembers[t] = make([]string, len(senderRingMembers[t-1]))
					senderRingMembers[t] = make([]string, len(recipientRingMembers[t-1]))
					copy(recipientRingMembers[t], senderRingMembers[t-1])
					copy(senderRingMembers[t], recipientRingMembers[t-1])
					payload.Recipient = txData.Payloads[t-1].Sender

					sendersPrivateKeys[t] = addresses.GenerateNewPrivateKey()
					var addr *addresses.Address
					if addr, err = sendersPrivateKeys[t].GenerateAddress(true, nil, 0, nil); err != nil {
						return
					}
					payload.Sender = addr.EncodeAddr()
					continue
				}
			}

			if senderRingMembers[t], recipientRingMembers[t], err = builder.createZetherRing(&payload.Sender, &payload.Recipient, payload.Asset, payload.RingConfiguration, dataStorage); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return nil, nil, nil, nil, nil, 0, nil, err
	}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		chainKernelHash = reader.Get("chainKernelHash")

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
					if acc, err = accs.GetAccount(addr.PublicKey); err != nil {
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
					hasRollovers[p.G1().String()] = acc != nil && acc.DelegatedStake != nil && acc.DelegatedStake.HasDelegatedStake()
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
			for _, ringMember := range senderRingMembers[t] {
				if err = addPoint(ringMember); err != nil {
					return
				}
			}
			for _, ringMember := range recipientRingMembers[t] {
				if err = addPoint(ringMember); err != nil {
					return
				}
			}

			transfers[t].WitnessIndexes = helpers.ShuffleArray_for_Zether(payload.RingConfiguration.RingSize)

			rings[t] = ring
		}
		statusCallback("Wallet Addresses Found")

		return
	}); err != nil {
		return nil, nil, nil, nil, nil, 0, nil, err
	}
	statusCallback("Balances checked")

	for t := range transfers {
		if sendersWalletAddresses[t] == nil {
			transfers[t].SenderDecryptedBalance = transfers[t].Amount
		} else {

			var err error
			if transfers[t].SenderDecryptedBalance, err = builder.wallet.DecryptBalanceByPublicKey(sendersWalletAddresses[t].PublicKey, sendersEncryptedBalances[t], transfers[t].Asset, false, 0, true, true, ctx, statusCallback); err != nil {
				return nil, nil, nil, nil, nil, 0, nil, err
			}

		}
		if transfers[t].SenderDecryptedBalance == 0 {
			return nil, nil, nil, nil, nil, 0, nil, errors.New("You have no funds")
		}
		if transfers[t].SenderDecryptedBalance < txData.Payloads[t].Amount {
			return nil, nil, nil, nil, nil, 0, nil, errors.New("Not enough funds")
		}
	}

	statusCallback("Balances decoded")

	return transfers, emap, hasRollovers, rings, publicKeyIndexes, chainHeight, chainKernelHash, nil
}

func (builder *TxsBuilder) CreateZetherTx(txData *TxBuilderCreateZetherTxData, propagateTx, awaitAnswer, awaitBroadcast bool, validateTx bool, ctx context.Context, statusCallback func(string)) (*transaction.Transaction, error) {

	builder.lock.Lock()
	defer builder.lock.Unlock()

	transfers, emap, hasRollovers, ringMembers, publicKeyIndexes, chainHeight, chainKernelHash, err := builder.prebuild(txData, ctx, statusCallback)
	if err != nil {
		return nil, err
	}

	feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		feesFinal[t] = payload.Fee.WizardTransactionFee
	}

	var tx *transaction.Transaction
	if tx, err = wizard.CreateZetherTx(transfers, emap, hasRollovers, ringMembers, chainHeight-1, chainKernelHash, publicKeyIndexes, feesFinal, validateTx, ctx, statusCallback); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMempool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL, ctx); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TxsBuilder) CreateForgingTransactions(blkComplete *block_complete.BlockComplete, forgerPublicKey []byte) (*transaction.Transaction, error) {

	gui.GUI.Info("CreateForgingTransactions 1")
	forger, err := addresses.CreateAddr(forgerPublicKey, nil, nil, 0, nil)
	if err != nil {
		return nil, err
	}

	reward := config_reward.GetRewardAt(blkComplete.Height)

	builder.lock.Lock()
	defer builder.lock.Unlock()

	//reward
	txData := &TxBuilderCreateZetherTxData{
		Payloads: []*TxBuilderCreateZetherTxPayload{
			{
				forger.EncodeAddr(),
				config_coins.NATIVE_ASSET_FULL,
				0,
				"",
				blkComplete.StakingAmount,
				&ZetherRingConfiguration{64, &ZetherSenderRingType{}, &ZetherRecipientRingType{}},
				nil,
				&wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, false}, false, 0, 0},
				&wizard.WizardZetherPayloadExtraStaking{},
			},
			{
				"",
				config_coins.NATIVE_ASSET_FULL,
				reward,
				forger.EncodeAddr(),
				0,
				&ZetherRingConfiguration{64, &ZetherSenderRingType{}, &ZetherRecipientRingType{}},
				nil,
				&wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, false}, false, 0, 0},
				&wizard.WizardZetherPayloadExtraStakingReward{nil, reward},
			},
		},
	}

	transfers, emap, hasRollovers, ringMembers, publicKeyIndexes, _, _, err := builder.prebuild(txData, context.Background(), func(string) {})
	if err != nil {
		return nil, err
	}

	gui.GUI.Info("CreateForgingTransactions 2")

	feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		feesFinal[t] = payload.Fee.WizardTransactionFee
	}

	chainHeight := blkComplete.Height
	if chainHeight > 0 {
		chainHeight--
	}

	var tx *transaction.Transaction
	if tx, err = wizard.CreateZetherTx(transfers, emap, hasRollovers, ringMembers, chainHeight, blkComplete.PrevKernelHash, publicKeyIndexes, feesFinal, false, context.Background(), func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("CreateForgingTransactions 3")

	return tx, nil
}
