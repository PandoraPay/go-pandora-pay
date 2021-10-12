package forging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type ForgingWallet struct {
	addresses             []*ForgingWalletAddress
	addressesMap          map[string]*ForgingWalletAddress
	workersAddresses      []int
	workers               []*ForgingWorkerThread
	updatePlainAccounts   *multicast.MulticastChannel
	updateWalletAddressCn chan *ForgingWalletAddressUpdate
	loadBalancesCn        chan struct{}
	workersCreatedCn      <-chan []*ForgingWorkerThread
	workersDestroyedCn    <-chan struct{}
	forging               *Forging
}

type ForgingWalletAddressUpdate struct {
	delegatedPriv helpers.HexBytes
	pubKey        helpers.HexBytes
}

type ForgingWalletAddress struct {
	delegatedPrivateKey     *addresses.PrivateKey
	delegatedStakePublicKey helpers.HexBytes //20 byte
	publicKey               helpers.HexBytes //20byte
	publicKeyStr            string
	plainAcc                *plain_account.PlainAccount
	workerIndex             int
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.delegatedPrivateKey,
		walletAddr.delegatedStakePublicKey,
		walletAddr.publicKey,
		walletAddr.publicKeyStr,
		walletAddr.plainAcc,
		walletAddr.workerIndex,
	}
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKey []byte) {
	w.updateWalletAddressCn <- &ForgingWalletAddressUpdate{
		delegatedPriv,
		pubKey,
	}
	return
}

func (w *ForgingWallet) RemoveWallet(DelegatedStakePublicKey []byte) { //20 byte
	w.AddWallet(nil, DelegatedStakePublicKey)
}

func (w *ForgingWallet) accountUpdated(addr *ForgingWalletAddress) {
	if addr.workerIndex != -1 {
		w.workers[addr.workerIndex].addWalletAddressCn <- addr.clone()
	}
}

func (w *ForgingWallet) accountInserted(addr *ForgingWalletAddress) {
	min := 0
	index := -1
	for i := 0; i < len(w.workersAddresses); i++ {
		if i == 0 || min > w.workersAddresses[i] {
			min = w.workersAddresses[i]
			index = i
		}
	}

	addr.workerIndex = index
	if index != -1 {
		w.workersAddresses[index]++
		w.workers[index].addWalletAddressCn <- addr.clone()
	}
}

func (w *ForgingWallet) accountRemoved(addr *ForgingWalletAddress) {
	if addr.workerIndex != -1 {
		w.workers[addr.workerIndex].removeWalletAddressCn <- addr.publicKeyStr
		w.workersAddresses[addr.workerIndex]--
		addr.workerIndex = -1
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
			for _, addr := range w.addresses {
				w.accountInserted(addr)
			}
		case _, ok := <-w.workersDestroyedCn:
			if !ok {
				return
			}
			w.workers = []*ForgingWorkerThread{}
			w.workersAddresses = []int{}
			for _, addr := range w.addresses {
				addr.workerIndex = -1
			}
		case update, ok := <-w.updateWalletAddressCn:
			if !ok {
				return
			}
			key := string(update.pubKey)

			//let's delete it
			if update.delegatedPriv == nil {

				if w.addressesMap[key] != nil {
					delete(w.addressesMap, key)
					for i, address := range w.addresses {
						if bytes.Equal(address.publicKey, update.pubKey) {
							w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
							w.accountRemoved(address)
							break
						}
					}
				}

			} else {

				delegatedPrivateKey := &addresses.PrivateKey{Key: update.delegatedPriv}
				delegatedStakePublicKey := delegatedPrivateKey.GeneratePublicKey()

				//let's read the balance
				if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

					chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

					plainAccs := plain_accounts.NewPlainAccounts(reader)

					var plainAcc *plain_account.PlainAccount
					if plainAcc, err = plainAccs.GetPlainAccount(update.pubKey, chainHeight); err != nil {
						return
					}
					if plainAcc == nil {
						return errors.New("Plain Account was not found")
					}

					if plainAcc.DelegatedStake == nil || !bytes.Equal(plainAcc.DelegatedStake.DelegatedStakePublicKey, delegatedStakePublicKey) {
						return errors.New("Delegated stake is not matching")
					}

					address := w.addressesMap[key]
					if address == nil {
						address = &ForgingWalletAddress{
							delegatedPrivateKey,
							delegatedStakePublicKey,
							update.pubKey,
							string(update.pubKey),
							plainAcc,
							-1,
						}
						w.addresses = append(w.addresses, address)
						w.addressesMap[key] = address
						w.accountInserted(address)
					} else {
						address.delegatedPrivateKey = delegatedPrivateKey
						address.delegatedStakePublicKey = delegatedStakePublicKey
						address.plainAcc = plainAcc
						w.accountUpdated(address)
					}

					return
				}); err != nil {
					gui.GUI.Error(err)
				}

			}
		case plainAccountData, ok := <-updatePlainAccountsCn:
			if !ok {
				return
			}

			plainAccounts := plainAccountData.(*plain_accounts.PlainAccounts)

			for k, v := range plainAccounts.HashMap.Committed {
				if w.addressesMap[k] != nil {
					if v.Stored == "update" {
						w.addressesMap[k].plainAcc = v.Element.(*plain_account.PlainAccount)
						w.accountUpdated(w.addressesMap[k])
					} else if v.Stored == "delete" {
						w.addressesMap[k].plainAcc = nil
						w.accountUpdated(w.addressesMap[k])
					}
				}
			}

		case _, ok := <-w.loadBalancesCn:
			if !ok {
				return
			}

			if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

				chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
				plainAccs := plain_accounts.NewPlainAccounts(reader)

				for _, address := range w.addresses {

					var plainAcc *plain_account.PlainAccount
					if plainAcc, err = plainAccs.GetPlainAccount(address.publicKey, chainHeight); err != nil {
						return
					}

					address.plainAcc = plainAcc
					w.accountUpdated(address)
				}

				return nil
			}); err != nil {
				gui.GUI.Error(err)
			}
		}

	}

}

func (w *ForgingWallet) loadBalances() {
	w.loadBalancesCn <- struct{}{}
}
