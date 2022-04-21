package forging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/tevino/abool"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address/shared_staked"
)

type ForgingWallet struct {
	addressesMap          map[string]*ForgingWalletAddress
	workersAddresses      []int
	workers               []*ForgingWorkerThread
	updateNewChainUpdate  *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates]
	updateWalletAddressCn chan *ForgingWalletAddressUpdate
	workersCreatedCn      <-chan []*ForgingWorkerThread
	workersDestroyedCn    <-chan struct{}
	forging               *Forging
	initialized           *abool.AtomicBool
}

type ForgingWalletAddressUpdate struct {
	chainHeight   uint64
	publicKeyHash []byte
	sharedStaked  *shared_staked.WalletAddressSharedStaked
	plainAcc      *plain_account.PlainAccount
}

func (w *ForgingWallet) AddWallet(publicKeyHash []byte, sharedStaked *shared_staked.WalletAddressSharedStaked, hasAccount bool, plainAcc *plain_account.PlainAccount, chainHeight uint64) (err error) {

	if !config_forging.FORGING_ENABLED || w.initialized.IsNotSet() {
		return
	}

	if !hasAccount {

		//let's read the balance
		if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

			chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
			dataStorage := data_storage.NewDataStorage(reader)

			if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(publicKeyHash); err != nil {
				return
			}

			return
		}); err != nil {
			return
		}

	}

	w.updateWalletAddressCn <- &ForgingWalletAddressUpdate{
		chainHeight,
		publicKeyHash,
		sharedStaked,
		plainAcc,
	}
	return
}

func (w *ForgingWallet) RemoveWallet(publicKeyHash []byte, hasAccount bool, plainAcc *plain_account.PlainAccount, chainHeight uint64) { //20 byte
	w.AddWallet(publicKeyHash, nil, hasAccount, plainAcc, chainHeight)
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

	w.workers[addr.workerIndex].addWalletAddressCn <- addr
}

func (w *ForgingWallet) removeAccountFromForgingWorkers(publicKey string) {

	addr := w.addressesMap[publicKey]

	if addr != nil && addr.workerIndex != -1 {
		w.workers[addr.workerIndex].removeWalletAddressCn <- addr.publicKeyHashStr
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

			key := string(update.publicKeyHash)

			//let's delete it
			if update.sharedStaked == nil || update.sharedStaked.PrivateKey == nil {
				w.removeAccountFromForgingWorkers(key)
			} else {

				if err = func() (err error) {

					if update.plainAcc == nil {
						return errors.New("Account was not found")
					}

					if !update.plainAcc.DelegatedStake.HasDelegatedStake() || !bytes.Equal(update.plainAcc.DelegatedStake.DelegatedStakePublicKey, update.sharedStaked.PublicKey) {
						return errors.New("Delegated stake is not matching")
					}

					if update.plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
						return errors.New("DelegatedStakeFee is less than it should be")
					}

					stakingAvailable := update.plainAcc.StakeAvailable

					if stakingAvailable < config_stake.GetRequiredStake(update.chainHeight) {
						return errors.New("Your stake is not accepted because you will need at least the minimum staking amount")
					}

					address := w.addressesMap[key]
					if address == nil {

						address = &ForgingWalletAddress{
							update.publicKeyHash,
							string(update.publicKeyHash),
							update.sharedStaked.PrivateKey,
							update.sharedStaked.PublicKey,
							update.plainAcc.DelegatedStake.DelegatedStakeFee,
							stakingAvailable,
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

			chainHash = update.BlockHash

			for k, v := range update.PlainAccounts.HashMap.Committed {
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

							stakingAvailable := plainAcc.StakeAvailable

							if stakingAvailable < config_stake.GetRequiredStake(update.BlockHeight) {
								return errors.New("Your stake is not accepted because you will need at least the minimum staking amount")
							}

							w.addressesMap[k].delegatedStakeFee = plainAcc.DelegatedStake.DelegatedStakeFee
							w.addressesMap[k].stakingAvailable = stakingAvailable
							w.addressesMap[k].chainHash = chainHash

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
