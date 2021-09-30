package api_delegates_node

import (
	"bytes"
	"pandora-pay/addresses"
	plain_accounts "pandora-pay/blockchain/data_storage/plain-accounts"
	plain_account "pandora-pay/blockchain/data_storage/plain-accounts/plain-account"
	"pandora-pay/recovery"
	wallet_address "pandora-pay/wallet/address"
	"sync/atomic"
	"time"
)

func (api *APIDelegatesNode) execute() {
	recovery.SafeGo(func() {

		updateNewChainUpdateListener := api.chain.UpdateNewChain.AddListener()
		defer api.chain.UpdateNewChain.RemoveChannel(updateNewChainUpdateListener)

		for {
			data, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			chainHeight := data.(uint64)
			atomic.StoreUint64(&api.chainHeight, chainHeight)
		}
	})

	recovery.SafeGo(func() {

		lastHeight := uint64(0)
		api.ticker = time.NewTicker(10 * time.Second)

		for {
			if _, ok := <-api.ticker.C; !ok {
				return
			}

			chainHeight := atomic.LoadUint64(&api.chainHeight)
			if lastHeight != chainHeight {
				lastHeight = chainHeight

				api.pendingDelegatesStakesChanges.Range(func(key, value interface{}) bool {
					pendingDelegateStakeChange := value.(*apiPendingDelegateStakeChange)
					if chainHeight >= pendingDelegateStakeChange.blockHeight+10 {
						api.pendingDelegatesStakesChanges.Delete(key)
					}
					return true
				})
			}
		}
	})

	api.updateAccountsChanges()

}

func (api *APIDelegatesNode) updateAccountsChanges() {

	recovery.SafeGo(func() {

		updatePlainAccountsCn := api.chain.UpdatePlainAccounts.AddListener()
		defer api.chain.UpdatePlainAccounts.RemoveChannel(updatePlainAccountsCn)

		for {

			plainAccsData, ok := <-updatePlainAccountsCn
			if !ok {
				return
			}

			plainAccs := plainAccsData.(*plain_accounts.PlainAccounts)

			for k, v := range plainAccs.HashMap.Committed {
				data, loaded := api.pendingDelegatesStakesChanges.Load(k)
				if loaded {

					pendingDelegatingStakeChange := data.(*apiPendingDelegateStakeChange)

					if v.Stored == "update" {
						plainAcc := v.Element.(*plain_account.PlainAccount)
						if plainAcc.HasDelegatedStake() && bytes.Equal(plainAcc.DelegatedStake.DelegatedPublicKey, pendingDelegatingStakeChange.delegatePublicKey) {

							addr, err := addresses.CreateAddr(pendingDelegatingStakeChange.publicKey, nil, 0, nil)
							if err != nil {
								continue
							}

							_ = api.wallet.AddDelegateStakeAddress(&wallet_address.WalletAddress{
								wallet_address.VERSION_NORMAL,
								"Delegate Stake",
								0,
								false,
								nil,
								nil,
								pendingDelegatingStakeChange.publicKey,
								make(map[string]*wallet_address.WalletAddressBalanceDecoded),
								addr.EncodeAddr(),
								"",
								&wallet_address.WalletAddressDelegatedStake{
									&addresses.PrivateKey{Key: pendingDelegatingStakeChange.delegatePrivateKey.Key},
									pendingDelegatingStakeChange.delegatePublicKey,
									0,
								},
							}, true)
						}
					}

				}
			}
		}
	})

}
