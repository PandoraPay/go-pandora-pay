package forging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_nodes"
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
	updatePlainAccounts   *multicast.MulticastChannel[*plain_accounts.PlainAccounts]
	updateWalletAddressCn chan *ForgingWalletAddressUpdate
	workersCreatedCn      <-chan []*ForgingWorkerThread
	workersDestroyedCn    <-chan struct{}
	forging               *Forging
}

type ForgingWalletAddressUpdate struct {
	delegatedPriv helpers.HexBytes
	pubKey        helpers.HexBytes
	hasPlainAcc   bool
	plainAcc      *plain_account.PlainAccount
}

type ForgingWalletAddress struct {
	delegatedPrivateKey     *addresses.PrivateKey
	delegatedStakePublicKey helpers.HexBytes //20 byte
	delegatedStakeFee       uint64
	publicKey               helpers.HexBytes //20byte
	publicKeyStr            string
	plainAcc                *plain_account.PlainAccount
	workerIndex             int
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.delegatedPrivateKey,
		walletAddr.delegatedStakePublicKey,
		walletAddr.delegatedStakeFee,
		walletAddr.publicKey,
		walletAddr.publicKeyStr,
		walletAddr.plainAcc,
		walletAddr.workerIndex,
	}
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKey []byte, hasPlainAcc bool, plainAcc *plain_account.PlainAccount) {
	w.updateWalletAddressCn <- &ForgingWalletAddressUpdate{
		delegatedPriv,
		pubKey,
		hasPlainAcc,
		plainAcc,
	}
	return
}

func (w *ForgingWallet) RemoveWallet(DelegatedStakePublicKey []byte, hasPlainAcc bool, plainAcc *plain_account.PlainAccount) { //20 byte
	w.AddWallet(nil, DelegatedStakePublicKey, hasPlainAcc, plainAcc)
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

func (w *ForgingWallet) deleteAccount(publicKey []byte) {

	delete(w.addressesMap, string(publicKey))
	for i, address := range w.addresses {
		if bytes.Equal(address.publicKey, publicKey) {
			w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
			w.accountRemoved(address)
			return
		}
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

				w.deleteAccount(update.pubKey)

			} else {

				delegatedPrivateKey := &addresses.PrivateKey{Key: update.delegatedPriv}
				delegatedStakePublicKey := delegatedPrivateKey.GeneratePublicKey()

				plainAcc := update.plainAcc

				if err = func() (err error) {

					if !update.hasPlainAcc {
						//let's read the balance
						if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

							chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
							plainAccs := plain_accounts.NewPlainAccounts(reader)
							if plainAcc, err = plainAccs.GetPlainAccount(update.pubKey, chainHeight); err != nil {
								return
							}

							return
						}); err != nil {
							return
						}
					}

					if plainAcc == nil {
						return errors.New("Plain Account was not found")
					}

					if !plainAcc.DelegatedStake.HasDelegatedStake() || !bytes.Equal(plainAcc.DelegatedStake.DelegatedStakePublicKey, delegatedStakePublicKey) {
						return errors.New("Delegated stake is not matching")
					}

					if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
						return errors.New("DelegatedStakeFee is less than it should be")
					}

					if plainAcc.DelegatedStake.DelegatedStakeFee > 0 && len(config_nodes.DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY) == 0 {
						return errors.New("DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY argument is missing")
					}

					delegatedStakeFee := plainAcc.DelegatedStake.DelegatedStakeFee

					address := w.addressesMap[key]
					if address == nil {
						address = &ForgingWalletAddress{
							delegatedPrivateKey,
							delegatedStakePublicKey,
							delegatedStakeFee,
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
						address.delegatedStakeFee = delegatedStakeFee
						address.plainAcc = plainAcc
						w.accountUpdated(address)
					}

					return
				}(); err != nil {
					w.deleteAccount(update.pubKey)
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

							if !bytes.Equal(plainAcc.DelegatedStake.DelegatedStakePublicKey, w.addressesMap[k].delegatedStakePublicKey) {
								return errors.New("delegated public key is no longer matching")
							}

							if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
								return errors.New("DelegatedStakeFee is less than it should be")
							}

							if plainAcc.DelegatedStake.DelegatedStakeFee > 0 && len(config_nodes.DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY) == 0 {
								return errors.New("DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY argument is missing")
							}

							w.addressesMap[k].delegatedStakeFee = plainAcc.DelegatedStake.DelegatedStakeFee
							w.addressesMap[k].plainAcc = plainAcc

							w.accountUpdated(w.addressesMap[k])

							return
						}(); err != nil {
							w.deleteAccount([]byte(k))
							gui.GUI.Error(err)
						}

					} else if v.Stored == "delete" {
						w.deleteAccount([]byte(k))
						gui.GUI.Error("Account was deleted")
					}

				}
			}

		}

	}

}
