package forging

import (
	"encoding/binary"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_forging"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type ForgingWallet struct {
	addressesMap          map[string]*ForgingWalletAddress
	workersAddresses      []int
	workers               []*ForgingWorkerThread
	updatePlainAccounts   *multicast.MulticastChannel[*plain_accounts.PlainAccounts]
	updateWalletAddressCn chan *ForgingWalletAddressUpdate
	workersCreatedCn      <-chan []*ForgingWorkerThread
	workersDestroyedCn    <-chan struct{}
	forging               *Forging
}

type ForgingWalletAddressUpdate struct {
	chainHeight uint64
	privateKey  []byte
	publicKey   []byte
	plainAcc    *plain_account.PlainAccount
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKey []byte, hasPlainAcc bool, plainAcc *plain_account.PlainAccount, chainHeight uint64) (err error) {

	if !config_forging.FORGING_ENABLED {
		return
	}

	if !hasPlainAcc {

		//let's read the balance
		if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

			chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
			plainAccs := plain_accounts.NewPlainAccounts(reader)
			if plainAcc, err = plainAccs.GetPlainAccount(pubKey, chainHeight); err != nil {
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
		plainAcc,
	}
	return
}

func (w *ForgingWallet) RemoveWallet(DelegatedStakePublicKey []byte, hasPlainAcc bool, plainAcc *plain_account.PlainAccount, chainHeight uint64) { //20 byte
	w.AddWallet(nil, DelegatedStakePublicKey, hasPlainAcc, plainAcc, chainHeight)
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
	newWorkerIndex := addr.workerIndex

	if newWorkerIndex != -1 {
		w.workers[addr.workerIndex].addWalletAddressCn <- addr.clone()
	}

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

func (w *ForgingWallet) processUpdates() {

	var err error

	updatePlainAccountsCn := w.updatePlainAccounts.AddListener()
	defer w.updatePlainAccounts.RemoveChannel(updatePlainAccountsCn)

	for {
		select {
		case workers, ok := <-w.workersCreatedCn:
			if !ok {
				return
			}
			w.workers = workers
			w.workersAddresses = make([]int, len(workers))
			for _, addr := range w.addressesMap {
				w.updateAccountToForgingWorkers(addr)
			}
		case _, ok := <-w.workersDestroyedCn:
			if !ok {
				return
			}
			w.workers = []*ForgingWorkerThread{}
			w.workersAddresses = []int{}
			for _, addr := range w.addressesMap {
				addr.workerIndex = -1
			}
		case update, ok := <-w.updateWalletAddressCn:
			if !ok {
				return
			}
			key := string(update.publicKey)

			//let's delete it
			if update.privateKey == nil {
				w.removeAccountFromForgingWorkers(key)
			} else {

				plainAcc := update.plainAcc

				if err = func() (err error) {

					if plainAcc == nil {
						return errors.New("Plain Account was not found")
					}

					if !plainAcc.DelegatedStake.HasDelegatedStake() {
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
							plainAcc,
							-1,
						}
						w.addressesMap[key] = address
					} else {
						address.plainAcc = plainAcc
					}

					w.updateAccountToForgingWorkers(address)

					return
				}(); err != nil {
					w.deleteAccount(key)
					gui.GUI.Error(err)
				}

			}
		case plainAccounts, ok := <-updatePlainAccountsCn:
			if !ok {
				return
			}

			for k, v := range plainAccounts.HashMap.Committed {
				if w.addressesMap[k] != nil {
					if v.Stored == "update" {

						plainAcc := v.Element.(*plain_account.PlainAccount)

						if err = func() (err error) {

							if !plainAcc.DelegatedStake.HasDelegatedStake() {
								return errors.New("has no longer delegated stake")
							}

							w.addressesMap[k].plainAcc = plainAcc

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
