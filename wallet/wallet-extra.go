package wallet

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/recovery"
)

func (wallet *Wallet) updateAccountsChanges() {

	recovery.SafeGo(func() {
		var err error
		updateAccountsCn := wallet.updateAccounts.AddListener()
		defer wallet.updateAccounts.RemoveChannel(updateAccountsCn)

		for {
			accsData, ok := <-updateAccountsCn
			if !ok {
				return
			}

			accs := accsData.(*accounts.Accounts)

			wallet.Lock()
			for k, v := range accs.HashMap.Committed {
				if wallet.addressesMap[k] != nil {

					if v.Stored == "update" {
						acc := v.Element.(*account.Account)
						if err = wallet.refreshWallet(acc, wallet.addressesMap[k], false); err != nil {
							return
						}
					} else if v.Stored == "delete" {
						if err = wallet.refreshWallet(nil, wallet.addressesMap[k], false); err != nil {
							return
						}
					}

				}
			}
			wallet.Unlock()
		}
	})

}
