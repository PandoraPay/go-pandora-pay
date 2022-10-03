package txs_builder

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/exp/slices"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_coins"
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

func (builder *TxsBuilder) getRandomAccount(accs *accounts.Accounts, regs *registrations.Registrations) (addr *addresses.Address, acc *account.Account, reg *registration.Registration, err error) {

	if acc, err = accs.GetRandom(); err != nil {
		return nil, nil, nil, err
	}
	if acc == nil {
		return nil, nil, nil, errors.New("Error getting any random account")
	}

	if reg, err = regs.Get(string(acc.Key)); err != nil {
		return nil, nil, nil, err
	}

	if addr, err = addresses.CreateAddr(acc.Key, false, nil, nil, nil, 0, nil); err != nil {
		return nil, nil, nil, err
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

func (builder *TxsBuilder) createZetherRing(allAlreadyUsed map[string]bool, senderRing *[]string, recipientRing *[]string, sender, receiver *string, assetId []byte, ringConfiguration *ZetherRingConfiguration, hasRollovers map[string]bool, dataStorage *data_storage.DataStorage) (err error) {

	alreadyUsed := make(map[string]bool)
	var addr, addrtemp *addresses.Address

	for i := 0; i < len(*senderRing); i++ {
		if addr, err = addresses.DecodeAddr((*senderRing)[i]); err != nil {
			return
		}
		alreadyUsed[string(addr.PublicKey)] = true
	}
	for i := 0; i < len(*recipientRing); i++ {
		if addr, err = addresses.DecodeAddr((*recipientRing)[i]); err != nil {
			return
		}
		alreadyUsed[string(addr.PublicKey)] = true
	}

	var reg *registration.Registration

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
		return
	}

	setAddress := func(ring *[]string, address *string, requireStakedAccounts, avoidStakedAccounts bool) (err error) {
		if *address == "" {
			if accs.Count == uint64(len(alreadyUsed)) {
				return errors.New("Accounts have only member. Impossible to get random recipient")
			}
			for {
				if addr, _, reg, err = builder.getRandomAccount(accs, dataStorage.Regs); err != nil {
					return
				}
				if avoidStakedAccounts && reg.Staked {
					continue
				}
				if (requireStakedAccounts && !reg.Staked) || (!requireStakedAccounts && len(reg.SpendPublicKey) > 0) {
					continue
				}
				if alreadyUsed[string(addr.PublicKey)] || allAlreadyUsed[string(addr.PublicKey)] {
					continue
				}
				*address = addr.EncodeAddr()
				break
			}
		} else {
			if addr, err = addresses.DecodeAddr(*address); err != nil {
				return
			}
			var p *crypto.Point
			if p, err = addr.GetPoint(); err != nil {
				return
			}
			hasRollovers[p.String()] = addr.Staked
		}
		if alreadyUsed[string(addr.PublicKey)] {
			for i := range *ring {
				if addrtemp, err = addresses.DecodeAddr((*ring)[i]); err != nil {
					return
				}
				if bytes.Equal(addr.PublicKey, addrtemp.PublicKey) {
					if i == 0 {
						return
					}
					temp := (*ring)[0]
					(*ring)[0] = (*ring)[i]
					(*ring)[i] = temp
					return
				}
			}
			return errors.New("Address was used before")
		}
		*ring = append([]string{*address}, *ring...)
		alreadyUsed[string(addr.PublicKey)] = true
		allAlreadyUsed[string(addr.PublicKey)] = true
		return
	}

	includeMembers := func(ring *[]string, includeMembers []string) (err error) {
		for i := 0; i < len(includeMembers) && len(*ring) < ringConfiguration.RingSize/2; i++ {
			if addr, err = addresses.DecodeAddr(includeMembers[i]); err != nil {
				return
			}
			if alreadyUsed[string(addr.PublicKey)] {
				continue
			}
			alreadyUsed[string(addr.PublicKey)] = true
			allAlreadyUsed[string(addr.PublicKey)] = true
			*ring = append(*ring, addr.EncodeAddr())
		}
		return
	}

	newAccounts := func(ring *[]string, newAccounts int, avoidStakedAccounts bool) (err error) {
		for i := 0; i < newAccounts && len(*ring) < ringConfiguration.RingSize/2; i++ {
			priv := addresses.GenerateNewPrivateKey()

			staked := false
			if !avoidStakedAccounts && rand.Intn(100) < 10 {
				staked = true
			}

			if addr, err = priv.GenerateAddress(staked, nil, true, nil, 0, nil); err != nil {
				return
			}

			alreadyUsed[string(addr.PublicKey)] = true
			allAlreadyUsed[string(addr.PublicKey)] = true
			hasRollovers[priv.GeneratePublicKeyPoint().String()] = staked

			*ring = append(*ring, addr.EncodeAddr())
		}
		return
	}

	newRandomAccounts := func(ring *[]string, requireStakedAccounts, avoidStakedAccounts bool) (err error) {

		for len(*ring) < ringConfiguration.RingSize/2 {

			if accs.Count <= uint64(len(alreadyUsed)) {
				priv := addresses.GenerateNewPrivateKey()
				if addr, err = priv.GenerateAddress(requireStakedAccounts, nil, true, nil, 0, nil); err != nil {
					return
				}
			} else {
				if addr, _, reg, err = builder.getRandomAccount(accs, dataStorage.Regs); err != nil {
					return
				}
				if alreadyUsed[string(addr.PublicKey)] || allAlreadyUsed[string(addr.PublicKey)] {
					continue
				}
				if avoidStakedAccounts && reg.Staked {
					continue
				}
				if (requireStakedAccounts && !reg.Staked) || (!requireStakedAccounts && len(reg.SpendPublicKey) > 0) {
					continue
				}
				alreadyUsed[string(addr.PublicKey)] = true
				allAlreadyUsed[string(addr.PublicKey)] = true
			}

			*ring = append(*ring, addr.EncodeAddr())
		}

		return
	}

	if err = setAddress(senderRing, sender, ringConfiguration.SenderRingType.RequireStakedAccounts, ringConfiguration.SenderRingType.AvoidStakedAccounts); err != nil {
		return
	}
	if err = setAddress(recipientRing, receiver, ringConfiguration.RecipientRingType.RequireStakedAccounts, ringConfiguration.RecipientRingType.AvoidStakedAccounts); err != nil {
		return
	}

	if err = includeMembers(senderRing, ringConfiguration.SenderRingType.IncludeMembers); err != nil {
		return
	}
	if err = includeMembers(recipientRing, ringConfiguration.RecipientRingType.IncludeMembers); err != nil {
		return
	}

	if err = newAccounts(senderRing, ringConfiguration.SenderRingType.NewAccounts, ringConfiguration.SenderRingType.AvoidStakedAccounts); err != nil {
		return
	}
	if err = newAccounts(recipientRing, ringConfiguration.RecipientRingType.NewAccounts, ringConfiguration.RecipientRingType.AvoidStakedAccounts); err != nil {
		return
	}

	if err = newRandomAccounts(senderRing, ringConfiguration.SenderRingType.RequireStakedAccounts, ringConfiguration.SenderRingType.AvoidStakedAccounts); err != nil {
		return
	}
	if err = newRandomAccounts(recipientRing, ringConfiguration.RecipientRingType.RequireStakedAccounts, ringConfiguration.RecipientRingType.AvoidStakedAccounts); err != nil {
		return
	}

	return
}

func (builder *TxsBuilder) prebuild(txData *TxBuilderCreateZetherTxData, pendingTxs []*transaction.Transaction, blockHeight uint64, prevKernelHash []byte, ctx context.Context, statusCallback func(string)) ([]*wizard.WizardZetherTransfer, map[string]map[string][]byte, map[string]bool, [][]*bn256.G1, [][]*bn256.G1, map[string]*wizard.WizardZetherPublicKeyIndex, uint64, []byte, error) {

	sendersPrivateKeys := make([]*addresses.PrivateKey, len(txData.Payloads))
	sendersWalletAddresses := make([]*wallet_address.WalletAddress, len(txData.Payloads))
	sendAssets := make([][]byte, len(txData.Payloads))

	hasRollovers := make(map[string]bool)

	for t, payload := range txData.Payloads {

		if payload.Asset == nil {
			payload.Asset = config_coins.NATIVE_ASSET_FULL
		}
		if payload.Data == nil {
			payload.Data = &wizard.WizardTransactionData{[]byte{}, false}
		}
		if payload.RingConfiguration == nil {
			payload.RingConfiguration = &ZetherRingConfiguration{-1, &ZetherSenderRingType{false, false, nil, 0}, &ZetherRecipientRingType{false, false, nil, 0}}
		}
		if payload.Fee == nil {
			payload.Fee = &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}
		}

		sendAssets[t] = payload.Asset
		if payload.Sender == "" {

			sendersPrivateKeys[t] = addresses.GenerateNewPrivateKey()
			addr, err := sendersPrivateKeys[t].GenerateAddress(false, nil, true, nil, 0, nil)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, 0, nil, err
			}
			payload.Sender = addr.EncodeAddr()

		} else {

			addr, err := builder.wallet.GetWalletAddressByEncodedAddress(payload.Sender, true)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, 0, nil, err
			}

			if addr.PrivateKey == nil {
				return nil, nil, nil, nil, nil, nil, 0, nil, errors.New("Can't be used for transactions as the private key is missing")
			}

			if sendersPrivateKeys[t], err = addresses.NewPrivateKey(addr.PrivateKey.Key); err != nil {
				return nil, nil, nil, nil, nil, nil, 0, nil, err
			}
			sendersWalletAddresses[t] = addr

		}

	}

	senderRingMembers := make([][]string, len(txData.Payloads))
	recipientRingMembers := make([][]string, len(txData.Payloads))

	transfers := make([]*wizard.WizardZetherTransfer, len(txData.Payloads))
	emap := wizard.InitializeEmap(sendAssets)

	ringsSenderMembers := make([][]*bn256.G1, len(txData.Payloads))
	ringsRecipientMembers := make([][]*bn256.G1, len(txData.Payloads))
	for i := range txData.Payloads {
		ringsSenderMembers[i] = make([]*bn256.G1, 0)
		ringsRecipientMembers[i] = make([]*bn256.G1, 0)
	}
	publicKeyIndexes := make(map[string]*wizard.WizardZetherPublicKeyIndex)

	sendersEncryptedBalances := make([][]byte, len(txData.Payloads))

	for _, payload := range txData.Payloads {
		if err := builder.presetZetherRing(payload.RingConfiguration); err != nil {
			return nil, nil, nil, nil, nil, nil, 0, nil, err
		}
	}

	allAlreadyUsed := make(map[string]bool) //avoid having same decoy twice

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		for t, payload := range txData.Payloads {

			senderRingMembers[t] = make([]string, 0)
			recipientRingMembers[t] = make([]string, 0)

			if len(payload.WitnessIndexes) == 0 {
				payload.WitnessIndexes = helpers.ShuffleArray_for_Zether(payload.RingConfiguration.RingSize)
			}

			if payload.Extra != nil {
				switch payload.Extra.(type) {
				case *wizard.WizardZetherPayloadExtraStakingReward:
					recipientRingMembers[t] = append(recipientRingMembers[t], senderRingMembers[t-1]...)
					senderRingMembers[t] = append(senderRingMembers[t], recipientRingMembers[t-1]...)
					payload.Recipient = txData.Payloads[t-1].Sender

					sendersPrivateKeys[t] = addresses.GenerateNewPrivateKey()
					var addr *addresses.Address
					if addr, err = sendersPrivateKeys[t].GenerateAddress(false, nil, true, nil, 0, nil); err != nil {
						return
					}
					payload.Sender = addr.EncodeAddr()
					senderRingMembers[t][0] = payload.Sender

					payload.WitnessIndexes = slices.Clone(txData.Payloads[t-1].WitnessIndexes)
					aux := payload.WitnessIndexes[0]
					payload.WitnessIndexes[0] = payload.WitnessIndexes[1]
					payload.WitnessIndexes[1] = aux

					continue
				}
			}

			copyRingConfiguration := func(ringMembers [][]string, copyRingMembers int, ringType int) (err error) {
				if copyRingMembers == -1 {
					return
				}

				if t == 0 || payload.RingConfiguration.RingSize != txData.Payloads[copyRingMembers].RingConfiguration.RingSize {
					return fmt.Errorf("ring size needs to be identical for payloads %d and %d", t-1, t)
				}

				ringMembers[t] = append(ringMembers[t], ringMembers[copyRingMembers]...)

				permutation := make([]int, len(payload.WitnessIndexes)/2)
				for i := range payload.WitnessIndexes {
					if i%2 == ringType {
						payload.WitnessIndexes[i] = txData.Payloads[copyRingMembers].WitnessIndexes[i]
					} else {
						permutation[i/2] = txData.Payloads[copyRingMembers].WitnessIndexes[i]
					}
				}

				for {
					permutationIndex := helpers.ShuffleArray(len(permutation))
					for i := range payload.WitnessIndexes {
						if i%2 != ringType {
							payload.WitnessIndexes[i] = permutation[permutationIndex[i/2]]
						}
					}
					if payload.WitnessIndexes[0]%2 != payload.WitnessIndexes[1]%2 {
						break
					}

				}
				return
			}

			copySenderRing := -1
			copyRecipientRing := -1
			for i := 0; i < t; i++ {
				if copySenderRing == -1 && txData.Payloads[i].Sender == payload.Sender {
					copySenderRing = i
				}
				if copyRecipientRing == -1 && txData.Payloads[i].Recipient == payload.Recipient {
					copyRecipientRing = i
				}
				if txData.Payloads[i].Sender == payload.Sender && txData.Payloads[i].Recipient == payload.Recipient {
					return errors.New("Sender and Recipient rings would be identical and leak the sender and receiver")
				}
			}

			if err = copyRingConfiguration(senderRingMembers, copySenderRing, 0); err != nil {
				return
			}
			if err = copyRingConfiguration(recipientRingMembers, copyRecipientRing, 1); err != nil {
				return
			}

			if err = builder.createZetherRing(allAlreadyUsed, &senderRingMembers[t], &recipientRingMembers[t], &payload.Sender, &payload.Recipient, payload.Asset, payload.RingConfiguration, hasRollovers, dataStorage); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return nil, nil, nil, nil, nil, nil, 0, nil, err
	}

	var chainHeight uint64
	var chainKernelHash []byte

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		chainKernelHash = reader.Get("chainKernelHash")

		if chainHeight > 0 && prevKernelHash != nil && !bytes.Equal(chainKernelHash, prevKernelHash) {
			return errors.New("KernelHash is already too old")
		}

		for t, payload := range txData.Payloads {

			var accs *accounts.Accounts
			if accs, err = dataStorage.AccsCollection.GetMap(payload.Asset); err != nil {
				return
			}

			if !bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) && payload.Fee.Auto {
				var assetFeeLiquidity *asset_fee_liquidity.AssetFeeLiquidity
				if assetFeeLiquidity, err = dataStorage.GetAssetFeeLiquidityTop(payload.Asset); err != nil {
					return
				}
				if assetFeeLiquidity == nil {
					return errors.New("There is no Asset Fee Liquidity for this asset")
				}
				payload.Fee.Rate = assetFeeLiquidity.Rate
				payload.Fee.LeadingZeros = assetFeeLiquidity.LeadingZeros
			}

			transfers[t] = &wizard.WizardZetherTransfer{
				Asset:            payload.Asset,
				SenderPrivateKey: sendersPrivateKeys[t].Key[:],
				Recipient:        payload.Recipient,
				Amount:           payload.Amount,
				Burn:             payload.Burn,
				Data:             payload.Data,
				FeeRate:          payload.Fee.Rate,
				FeeLeadingZeros:  payload.Fee.LeadingZeros,
				PayloadExtra:     payload.Extra,
				WitnessIndexes:   payload.WitnessIndexes,
			}

			//parity := transfers[t].WitnessIndexes[0]%2 == 0

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
					return errors.New("A ring member was detected twice")
				}
				uniqueMap[string(addr.PublicKey)] = true

				if p, err = addr.GetPoint(); err != nil {
					return
				}

				var reg *registration.Registration
				if reg, err = dataStorage.Regs.Get(string(addr.PublicKey)); err != nil {
					return
				}

				if emap[string(payload.Asset)][p.G1().String()] == nil {

					var acc *account.Account
					if acc, err = accs.Get(string(addr.PublicKey)); err != nil {
						return
					}

					hasRollover := reg != nil && bytes.Equal(payload.Asset, config_coins.NATIVE_ASSET_FULL) && reg.Staked
					if hasRollover {
						hasRollovers[p.G1().String()] = hasRollover
					}

					var newBalance *crypto.ElGamal
					if acc != nil {
						newBalance = acc.Balance.Amount
					}

					if newBalance, err = wizard.GetZetherBalance(addr.PublicKey, newBalance, payload.Asset, hasRollovers[p.G1().String()], pendingTxs); err != nil {
						return
					}

					if newBalance != nil {
						emap[string(payload.Asset)][p.G1().String()] = newBalance.Serialize()
					}

					if isSender { //sender
						if newBalance != nil {
							sendersEncryptedBalances[t] = newBalance.Serialize()
						} else {
							sendersEncryptedBalances[t] = crypto.ConstructElGamal(p.G1(), crypto.ElGamal_BASE_G).Serialize()
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
							publicKeyIndex.RegistrationStaked = addr.Staked
							publicKeyIndex.RegistrationSpendPublicKey = addr.SpendPublicKey
							publicKeyIndex.RegistrationSignature = addr.Registration
						}
					}
				}

				if sender {
					if reg != nil && len(reg.SpendPublicKey) > 0 && payload.Extra == nil {
						transfers[t].SenderSpendRequired = true
						if sendersWalletAddresses[t].SpendPrivateKey == nil {
							return errors.New("Spend Private Key is missing")
						}
						if !bytes.Equal(sendersWalletAddresses[t].SpendPublicKey, reg.SpendPublicKey) {
							return errors.New("Wallet Spend Public Key is not matching")
						}
						transfers[t].SenderSpendPrivateKey = sendersWalletAddresses[t].SpendPrivateKey.Key
					}
				}

				if sender {
					ringSender = append(ringSender, p.G1())
				} else {
					ringRecipient = append(ringRecipient, p.G1())
				}

				return
			}

			for i, ringMember := range senderRingMembers[t] {
				if err = addPoint(ringMember, true, i == 0); err != nil {
					return
				}
			}
			for _, ringMember := range recipientRingMembers[t] {
				if err = addPoint(ringMember, false, false); err != nil {
					return
				}
			}

			ringsSenderMembers[t] = ringSender
			ringsRecipientMembers[t] = ringRecipient
		}
		statusCallback("Wallet Addresses Found")

		return
	}); err != nil {
		return nil, nil, nil, nil, nil, nil, 0, nil, err
	}
	statusCallback("Balances checked")

	for t := range transfers {

		verify := true

		if sendersWalletAddresses[t] == nil {
			transfers[t].SenderDecryptedBalance = transfers[t].Amount
		} else if sendersEncryptedBalances[t] != nil {

			if txData.Payloads[t].DecryptedBalance > 0 { // in case it was specified to avoid getting stuck
				decrypted, err := builder.wallet.DecryptBalance(sendersWalletAddresses[t], sendersEncryptedBalances[t], transfers[t].Asset, true, txData.Payloads[t].DecryptedBalance, true, ctx, statusCallback)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, 0, nil, err
				}
				transfers[t].SenderDecryptedBalance = decrypted
			} else {
				decrypted, err := builder.wallet.DecryptBalance(sendersWalletAddresses[t], sendersEncryptedBalances[t], transfers[t].Asset, false, 0, true, ctx, statusCallback)
				if err != nil {
					return nil, nil, nil, nil, nil, nil, 0, nil, err
				}
				transfers[t].SenderDecryptedBalance = decrypted
			}
		} else {
			verify = false
		}

		if verify {
			if transfers[t].SenderDecryptedBalance == 0 {
				return nil, nil, nil, nil, nil, nil, 0, nil, errors.New("You have no funds")
			}

			if transfers[t].SenderDecryptedBalance < txData.Payloads[t].Amount {
				return nil, nil, nil, nil, nil, nil, 0, nil, errors.New("Not enough funds")
			}
		}
	}

	statusCallback("Balances decoded")

	return transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, publicKeyIndexes, chainHeight, chainKernelHash, nil
}

func (builder *TxsBuilder) CreateZetherTx(txData *TxBuilderCreateZetherTxData, pendingTxs []*transaction.Transaction, propagateTx, awaitAnswer, awaitBroadcast bool, validateTx bool, ctx context.Context, statusCallback func(string)) (*transaction.Transaction, error) {

	if pendingTxs == nil {
		pendingTxs = builder.mempool.Txs.GetTxsOnlyList()
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, publicKeyIndexes, chainHeight, chainKernelHash, err := builder.prebuild(txData, pendingTxs, 0, nil, ctx, statusCallback)
	if err != nil {
		return nil, err
	}

	feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		feesFinal[t] = payload.Fee.WizardTransactionFee
	}

	var tx *transaction.Transaction
	if tx, err = wizard.CreateZetherTx(transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, chainHeight-1, chainKernelHash, publicKeyIndexes, feesFinal, ctx, statusCallback); err != nil {
		return nil, err
	}

	if err = builder.txsValidator.MarkAsValidatedTx(tx); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMempool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL, ctx); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TxsBuilder) CreateForgingTransactions(blkComplete *block_complete.BlockComplete, forgerPublicKey []byte, decryptedBalance uint64, pendingTxs []*transaction.Transaction) (*transaction.Transaction, error) {

	if pendingTxs == nil {
		pendingTxs = builder.mempool.Txs.GetTxsOnlyList()
	}

	gui.GUI.Info("CreateForgingTransactions 1")
	forger, err := addresses.CreateAddr(forgerPublicKey, false, nil, nil, nil, 0, nil)
	if err != nil {
		return nil, err
	}

	_, finalForgerReward, err := blockchain_types.ComputeBlockReward(blkComplete.Height, pendingTxs)
	if err != nil {
		return nil, err
	}

	chainHeight := blkComplete.Height
	if chainHeight > 0 {
		chainHeight--
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	//reward
	txData := &TxBuilderCreateZetherTxData{
		Payloads: []*TxBuilderCreateZetherTxPayload{
			{
				forger.EncodeAddr(),
				config_coins.NATIVE_ASSET_FULL,
				0,
				decryptedBalance,
				"",
				blkComplete.StakingAmount,
				&ZetherRingConfiguration{64, &ZetherSenderRingType{true, false, nil, 0}, &ZetherRecipientRingType{true, false, nil, 0}},
				nil,
				&wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, false}, false, 0, 0},
				&wizard.WizardZetherPayloadExtraStaking{},
				nil,
			},
			{
				"",
				config_coins.NATIVE_ASSET_FULL,
				finalForgerReward,
				finalForgerReward, //reward will be the encrypted Balance
				forger.EncodeAddr(),
				0,
				&ZetherRingConfiguration{64, &ZetherSenderRingType{true, false, nil, 0}, &ZetherRecipientRingType{true, false, nil, 0}},
				nil,
				&wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, false}, false, 0, 0},
				&wizard.WizardZetherPayloadExtraStakingReward{nil, finalForgerReward},
				nil,
			},
		},
	}

	transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, publicKeyIndexes, _, _, err := builder.prebuild(txData, pendingTxs, blkComplete.Height, blkComplete.PrevKernelHash, context.Background(), func(string) {})
	if err != nil {
		return nil, err
	}

	gui.GUI.Info("CreateForgingTransactions 2")

	feesFinal := make([]*wizard.WizardTransactionFee, len(txData.Payloads))
	for t, payload := range txData.Payloads {
		feesFinal[t] = payload.Fee.WizardTransactionFee
	}

	var tx *transaction.Transaction
	if tx, err = wizard.CreateZetherTx(transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, chainHeight, blkComplete.PrevKernelHash, publicKeyIndexes, feesFinal, context.Background(), func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("CreateForgingTransactions 3")

	if err = builder.txsValidator.MarkAsValidatedTx(tx); err != nil {
		return nil, err
	}

	//if err = builder.txsValidator.ValidateTx(tx); err != nil {
	//	return nil, err
	//}

	gui.GUI.Info("CreateForgingTransactions 4")

	return tx, nil
}
