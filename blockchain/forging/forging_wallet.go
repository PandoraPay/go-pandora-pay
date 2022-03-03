package forging

import (
	"context"
	"encoding/binary"
	"errors"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_forging"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"time"
)

type ForgingWallet struct {
	addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor
	addressesMap            map[string]*ForgingWalletAddress
	workersAddresses        []int
	workers                 []*ForgingWorkerThread
	updateAccounts          *multicast.MulticastChannel[*accounts.AccountsCollection]
	updateWalletAddressCn   chan *ForgingWalletAddressUpdate
	workersCreatedCn        <-chan []*ForgingWorkerThread
	workersDestroyedCn      <-chan struct{}
	decryptBalancesUpdates  *generics.Map[string, *ForgingWalletAddress]
	forging                 *Forging
}

type ForgingWalletAddressUpdate struct {
	chainHeight uint64
	privateKey  []byte
	publicKey   []byte
	account     *account.Account
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKey []byte, hasAccount bool, account *account.Account, chainHeight uint64) (err error) {

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

			if account, err = accs.GetAccount(pubKey); err != nil {
				return
			}

			return
		}); err != nil {
			return
		}

	}

	w.updateWalletAddressCn <- &ForgingWalletAddressUpdate{
		chainHeight,
		delegatedPriv,
		pubKey,
		account,
	}
	return
}

func (w *ForgingWallet) RemoveWallet(DelegatedStakePublicKey []byte, hasAccount bool, acc *account.Account, chainHeight uint64) { //20 byte
	w.AddWallet(nil, DelegatedStakePublicKey, hasAccount, acc, chainHeight)
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
			stakingAmountEncryptedBalance := addr.account.Balance.Amount
			stakingAmountEncryptedBalanceSerialized := stakingAmountEncryptedBalance.Serialize()
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

	updateAccountsCn := w.updateAccounts.AddListener()
	defer w.updateAccounts.RemoveChannel(updateAccountsCn)

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
			if update.privateKey == nil {
				w.removeAccountFromForgingWorkers(key)
			} else {

				if err = func() (err error) {

					if update.account == nil {
						return errors.New("Account was not found")
					}

					if !update.account.DelegatedStake.HasDelegatedStake() {
						return errors.New("Delegated stake is not matching")
					}

					address := w.addressesMap[key]
					if address == nil {

						keyPoint := new(crypto.BNRed).SetBytes(update.privateKey)

						address = &ForgingWalletAddress{
							&addresses.PrivateKey{Key: update.privateKey},
							keyPoint.BigInt(),
							update.publicKey,
							string(update.publicKey),
							update.account,
							0,
							-1,
						}
						w.addressesMap[key] = address
					} else {
						address.account = update.account
					}

					w.updateAccountToForgingWorkers(address)

					return
				}(); err != nil {
					w.deleteAccount(key)
					gui.GUI.Error(err)
				}

			}
		case accsCollection := <-updateAccountsCn:

			accs, _ := accsCollection.GetOnlyMap(config_coins.NATIVE_ASSET_FULL)
			if accs == nil {
				continue
			}

			for k, v := range accs.HashMap.Committed {
				if w.addressesMap[k] != nil {
					if v.Stored == "update" {

						acc := v.Element.(*account.Account)

						if err = func() (err error) {

							if !acc.DelegatedStake.HasDelegatedStake() {
								return errors.New("has no longer delegated stake")
							}

							w.addressesMap[k].account = acc
							w.updateAccountToForgingWorkers(w.addressesMap[k])

							return
						}(); err != nil {
							w.deleteAccount(k)
							gui.GUI.Error(err)
						}

					} else if v.Stored == "delete" {
						w.deleteAccount(k)
						gui.GUI.Error("Account was deleted from Forging")
					}

				}
			}

		}

	}

}
