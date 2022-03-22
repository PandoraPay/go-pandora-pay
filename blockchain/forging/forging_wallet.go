package forging

import (
	"context"
	"encoding/binary"
	"errors"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_forging"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"time"
)

type ForgingWallet struct {
	addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor
	addressesMap            map[string]*ForgingWalletAddress
	workersAddresses        []int
	workers                 []*ForgingWorkerThread
	updateNewChainUpdate    *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates]
	updateWalletAddressCn   chan *ForgingWalletAddressUpdate
	workersCreatedCn        <-chan []*ForgingWorkerThread
	workersDestroyedCn      <-chan struct{}
	decryptBalancesUpdates  *generics.Map[string, *ForgingWalletAddress]
	forging                 *Forging
}

type ForgingWalletAddressUpdate struct {
	chainHeight  uint64
	publicKey    []byte
	sharedStaked *wallet_address.WalletAddressSharedStaked
	account      *account.Account
	registration *registration.Registration
}

func (w *ForgingWallet) AddWallet(publicKey []byte, sharedStaked *wallet_address.WalletAddressSharedStaked, hasAccount bool, account *account.Account, reg *registration.Registration, chainHeight uint64) (err error) {

	if !config_forging.FORGING_ENABLED {
		return
	}

	if !hasAccount {

		//let's read the balance
		if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

			chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
			dataStorage := data_storage.NewDataStorage(reader)

			var accs *accounts.Accounts
			if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
				return
			}

			if account, err = accs.GetAccount(publicKey); err != nil {
				return
			}
			if reg, err = dataStorage.Regs.GetRegistration(publicKey); err != nil {
				return
			}

			return
		}); err != nil {
			return
		}

	}

	w.updateWalletAddressCn <- &ForgingWalletAddressUpdate{
		chainHeight,
		publicKey,
		sharedStaked,
		account,
		reg,
	}
	return
}

func (w *ForgingWallet) RemoveWallet(publicKey []byte, hasAccount bool, acc *account.Account, reg *registration.Registration, chainHeight uint64) { //20 byte
	w.AddWallet(publicKey, nil, hasAccount, acc, reg, chainHeight)
}

func (w *ForgingWallet) runDecryptBalanceAndNotifyWorkers() {

	var addr *ForgingWalletAddress
	for {

		found := false
		w.decryptBalancesUpdates.Range(func(publicKey string, _ *ForgingWalletAddress) bool {
			addr, _ = w.decryptBalancesUpdates.LoadAndDelete(publicKey)
			found = true
			return false
		})

		if !found {
			time.Sleep(10 * time.Millisecond)
			continue
		} else {
			stakingAmountEncryptedBalanceSerialized := addr.account.Balance.Amount.Serialize()
			addr.decryptedStakingBalance, _ = w.addressBalanceDecryptor.DecryptBalance("staking", addr.publicKey, addr.privateKey.Key, stakingAmountEncryptedBalanceSerialized, config_coins.NATIVE_ASSET_FULL, false, 0, true, context.Background(), func(string) {})

			w.workers[addr.workerIndex].addWalletAddressCn <- addr
		}
	}

}

func (w *ForgingWallet) updateAccountToForgingWorkers(addr *ForgingWalletAddress) {

	if len(w.workers) == 0 { //in case it was not started yet
		return
	}

	if addr.workerIndex == -1 {
		min := 0
		index := -1
		for i := 0; i < len(w.workersAddresses); i++ {
			if i == 0 || min > w.workersAddresses[i] {
				min = w.workersAddresses[i]
				index = i
			}
		}

		addr.workerIndex = index
		w.workersAddresses[index]++

	}

	w.decryptBalancesUpdates.Store(addr.publicKeyStr, addr.clone())
}

func (w *ForgingWallet) removeAccountFromForgingWorkers(publicKey string) {

	addr := w.addressesMap[publicKey]

	if addr != nil && addr.workerIndex != -1 {
		w.workers[addr.workerIndex].removeWalletAddressCn <- addr.publicKeyStr
		w.workersAddresses[addr.workerIndex]--
		addr.workerIndex = -1
	}
}

func (w *ForgingWallet) deleteAccount(publicKey string) {
	if addr := w.addressesMap[publicKey]; addr != nil {
		w.removeAccountFromForgingWorkers(publicKey)
	}
}

func (w *ForgingWallet) runProcessUpdates() {

	var err error

	updateNewChainCn := w.updateNewChainUpdate.AddListener()
	defer w.updateNewChainUpdate.RemoveChannel(updateNewChainCn)

	var chainHash []byte

	for {
		select {
		case workers := <-w.workersCreatedCn:

			w.workers = workers
			w.workersAddresses = make([]int, len(workers))
			for _, addr := range w.addressesMap {
				w.updateAccountToForgingWorkers(addr)
			}
		case <-w.workersDestroyedCn:

			w.workers = []*ForgingWorkerThread{}
			w.workersAddresses = []int{}
			for _, addr := range w.addressesMap {
				addr.workerIndex = -1
			}
		case update := <-w.updateWalletAddressCn:

			key := string(update.publicKey)

			//let's delete it
			if update.sharedStaked == nil || update.sharedStaked.PrivateKey == nil {
				w.removeAccountFromForgingWorkers(key)
			} else {

				if err = func() (err error) {

					if update.account == nil {
						return errors.New("Account was not found")
					}
					if update.registration == nil {
						return errors.New("Registration was not found")
					}

					if !update.registration.Staked {
						return errors.New("It is no longer staked")
					}

					address := w.addressesMap[key]
					if address == nil {

						keyPoint := new(crypto.BNRed).SetBytes(update.sharedStaked.PrivateKey.Key)

						address = &ForgingWalletAddress{
							update.sharedStaked.PrivateKey,
							keyPoint.BigInt(),
							update.publicKey,
							string(update.publicKey),
							update.account,
							0,
							-1,
							chainHash,
						}
						w.addressesMap[key] = address
						w.updateAccountToForgingWorkers(address)
					}

					return
				}(); err != nil {
					w.deleteAccount(key)
					gui.GUI.Error(err)
				}

			}
		case update := <-updateNewChainCn:

			accs, _ := update.AccsCollection.GetOnlyMap(config_coins.NATIVE_ASSET_FULL)
			if accs == nil {
				continue
			}

			chainHash = update.BlockHash

			for k, v := range update.Registrations.Committed {
				if w.addressesMap[k] != nil {
					if v.Stored == "update" {
						reg := v.Element.(*registration.Registration)
						if !reg.Staked {
							w.deleteAccount(k)
						}
					} else if v.Stored == "delete" {
						w.deleteAccount(k)
					}
				}
			}

			for k, v := range accs.HashMap.Committed {
				if w.addressesMap[k] != nil {
					if v.Stored == "update" {

						acc := v.Element.(*account.Account)

						w.addressesMap[k].account = acc
						w.addressesMap[k].chainHash = chainHash
						w.updateAccountToForgingWorkers(w.addressesMap[k])

					} else if v.Stored == "delete" {
						w.deleteAccount(k)
						gui.GUI.Error("Account was deleted from Forging")
					}

				}
			}

		}

	}

}
